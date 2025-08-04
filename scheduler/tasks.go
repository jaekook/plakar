package scheduler

import (
	"fmt"
	"time"

	"github.com/PlakarKorp/kloset/encryption"
	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/kloset/storage"
	"github.com/PlakarKorp/kloset/versioning"
	"github.com/PlakarKorp/plakar/agent"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/locate"
	"github.com/PlakarKorp/plakar/subcommands"
	"github.com/PlakarKorp/plakar/subcommands/backup"
	"github.com/PlakarKorp/plakar/subcommands/check"
	"github.com/PlakarKorp/plakar/subcommands/maintenance"
	"github.com/PlakarKorp/plakar/subcommands/restore"
	"github.com/PlakarKorp/plakar/subcommands/rm"
	"github.com/PlakarKorp/plakar/subcommands/sync"
)

func loadRepository(newCtx *appcontext.AppContext, name string) (*repository.Repository, storage.Store, error) {
	if err := newCtx.ReloadConfig(); err != nil {
		return nil, nil, fmt.Errorf("could not load configuration: %w", err)
	}

	storeConfig, err := newCtx.Config.GetRepository(name)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to get repository configuration: %w", err)
	}

	store, config, err := storage.Open(newCtx.GetInner(), storeConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("unable to open storage: %w", err)
	}

	repoConfig, err := storage.NewConfigurationFromWrappedBytes(config)
	if err != nil {
		store.Close(newCtx)
		return nil, nil, fmt.Errorf("unable to read repository configuration: %w", err)
	}

	if repoConfig.Version != versioning.FromString(storage.VERSION) {
		store.Close(newCtx)
		return nil, nil, fmt.Errorf("incompatible repository version: %s != %s", repoConfig.Version, storage.VERSION)
	}

	if passphrase, ok := storeConfig["passphrase"]; ok {
		key, err := encryption.DeriveKey(repoConfig.Encryption.KDFParams, []byte(passphrase))
		if err != nil {
			store.Close(newCtx)
			return nil, nil, fmt.Errorf("error deriving key: %w", err)
		}
		if !encryption.VerifyCanary(repoConfig.Encryption, key) {
			store.Close(newCtx)
			return nil, nil, fmt.Errorf("invalid passphrase")
		}
		newCtx.SetSecret(key)
	}

	repo, err := repository.NewNoRebuild(newCtx.GetInner(), newCtx.GetSecret(), store, config)
	if err != nil {
		store.Close(newCtx)
		return nil, store, fmt.Errorf("unable to open repository: %w", err)
	}
	return repo, store, nil
}

func (s *Scheduler) backupTask(taskset Task, task BackupConfig) {
	backupSubcommand := &backup.Backup{}
	backupSubcommand.Flags = subcommands.AgentSupport
	backupSubcommand.Silent = true
	backupSubcommand.Job = taskset.Name
	backupSubcommand.Path = task.Path
	backupSubcommand.Quiet = true
	backupSubcommand.Opts = make(map[string]string)
	if task.Check.Enabled {
		backupSubcommand.OptCheck = true
	}

	rmSubcommand := &rm.Rm{}
	rmSubcommand.Flags = subcommands.AgentSupport
	rmSubcommand.LocateOptions = locate.NewDefaultLocateOptions()
	rmSubcommand.LocateOptions.Job = task.Name

	for {
		tick := time.After(task.Interval)
		select {
		case <-s.ctx.Done():
			return
		case <-tick:
			storeConfig, err := s.ctx.Config.GetRepository(taskset.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error getting repository config: %s", err)
				continue
			}

			repo, store, err := loadRepository(s.ctx, taskset.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error loading repository: %s", err)
				continue
			}

			if retval, err := agent.ExecuteRPC(s.ctx, []string{"backup"}, backupSubcommand, storeConfig); err != nil || retval != 0 {
				s.ctx.GetLogger().Error("Error creating backup: %s", err)
				goto close
			}

			if task.Retention != 0 {
				rmSubcommand.LocateOptions.Before = time.Now().Add(-task.Retention)
				if retval, err := agent.ExecuteRPC(s.ctx, []string{"rm"}, rmSubcommand, storeConfig); err != nil || retval != 0 {
					s.ctx.GetLogger().Error("Error removing obsolete backups: %s", err)
					goto close
				}
			}

		close:
			repo.Close()
			store.Close(s.ctx)
		}
	}
}

func (s *Scheduler) checkTask(taskset Task, task CheckConfig) {
	checkSubcommand := &check.Check{}
	checkSubcommand.Flags = subcommands.AgentSupport
	checkSubcommand.LocateOptions = locate.NewDefaultLocateOptions()
	checkSubcommand.LocateOptions.Job = taskset.Name
	checkSubcommand.LocateOptions.Latest = task.Latest
	checkSubcommand.Silent = true
	if task.Path != "" {
		checkSubcommand.Snapshots = []string{":" + task.Path}
	}

	for {
		tick := time.After(task.Interval)
		select {
		case <-s.ctx.Done():
			return
		case <-tick:
			storeConfig, err := s.ctx.Config.GetRepository(taskset.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error getting repository config: %s", err)
				continue
			}

			repo, store, err := loadRepository(s.ctx, taskset.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error loading repository: %s", err)
				continue
			}

			retval, err := agent.ExecuteRPC(s.ctx, []string{"check"}, checkSubcommand, storeConfig)
			if err != nil || retval != 0 {
				s.ctx.GetLogger().Error("Error executing check: %s", err)
			}

			repo.Close()
			store.Close(s.ctx)
		}
	}
}

func (s *Scheduler) restoreTask(taskset Task, task RestoreConfig) {
	restoreSubcommand := &restore.Restore{}
	restoreSubcommand.Flags = subcommands.AgentSupport
	restoreSubcommand.OptJob = taskset.Name
	restoreSubcommand.Target = task.Target
	restoreSubcommand.Silent = true
	if task.Path != "" {
		restoreSubcommand.Snapshots = []string{":" + task.Path}
	}

	for {
		tick := time.After(task.Interval)
		select {
		case <-s.ctx.Done():
			return
		case <-tick:
			storeConfig, err := s.ctx.Config.GetRepository(taskset.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error getting repository config: %s", err)
				continue
			}

			repo, store, err := loadRepository(s.ctx, taskset.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error loading repository: %s", err)
				continue
			}

			retval, err := agent.ExecuteRPC(s.ctx, []string{"restore"}, restoreSubcommand, storeConfig)
			if err != nil || retval != 0 {
				s.ctx.GetLogger().Error("Error executing restore: %s", err)
			}

			repo.Close()
			store.Close(s.ctx)
		}
	}
}

func (s *Scheduler) syncTask(taskset Task, task SyncConfig) {
	syncSubcommand := &sync.Sync{}
	syncSubcommand.Flags = subcommands.AgentSupport
	syncSubcommand.PeerRepositoryLocation = task.Peer
	if task.Direction == SyncDirectionTo {
		syncSubcommand.Direction = "to"
	} else if task.Direction == SyncDirectionFrom {
		syncSubcommand.Direction = "from"
	} else if task.Direction == SyncDirectionWith {
		syncSubcommand.Direction = "with"
	} else {
		//return fmt.Errorf("invalid sync direction: %s", task.Direction)
		s.ctx.Cancel()
		return
	}
	//	if taskset.Repository.Passphrase != "" {
	//		syncSubcommand.DestinationRepositorySecret = []byte(taskset.Repository.Passphrase)
	//		_ = syncSubcommand.DestinationRepositorySecret

	//	syncSubcommand.OptJob = taskset.Name
	//	syncSubcommand.Target = task.Target
	//	syncSubcommand.Silent = true

	for {
		tick := time.After(task.Interval)
		select {
		case <-s.ctx.Done():
			return
		case <-tick:
			storeConfig, err := s.ctx.Config.GetRepository(taskset.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error getting repository config: %s", err)
				continue
			}

			repo, store, err := loadRepository(s.ctx, taskset.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error loading repository: %s", err)
				continue
			}

			retval, err := agent.ExecuteRPC(s.ctx, []string{"sync"}, syncSubcommand, storeConfig)
			if err != nil || retval != 0 {
				s.ctx.GetLogger().Error("sync: %s", err)
			} else {
				s.ctx.GetLogger().Info("sync: synchronization succeeded")
			}

			repo.Close()
			store.Close(s.ctx)
		}
	}
}

func (s *Scheduler) maintenanceTask(task MaintenanceConfig) {
	maintenanceSubcommand := &maintenance.Maintenance{}
	maintenanceSubcommand.Flags = subcommands.AgentSupport
	rmSubcommand := &rm.Rm{}
	rmSubcommand.Flags = subcommands.AgentSupport
	rmSubcommand.LocateOptions = locate.NewDefaultLocateOptions()
	rmSubcommand.LocateOptions.Job = "maintenance"

	for {
		tick := time.After(task.Interval)
		select {
		case <-s.ctx.Done():
			return
		case <-tick:
			storeConfig, err := s.ctx.Config.GetRepository(task.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error getting repository config: %s", err)
				continue
			}

			repo, store, err := loadRepository(s.ctx, task.Repository)
			if err != nil {
				s.ctx.GetLogger().Error("Error loading repository: %s", err)
				continue
			}

			retval, err := agent.ExecuteRPC(s.ctx, []string{"maintenance"}, maintenanceSubcommand, storeConfig)
			if err != nil || retval != 0 {
				s.ctx.GetLogger().Error("Error executing maintenance: %s", err)
				goto close
			} else {
				s.ctx.GetLogger().Info("maintenance of repository %s succeeded", task.Repository)
			}

			if task.Retention != 0 {
				rmSubcommand.LocateOptions.Before = time.Now().Add(-task.Retention)
				retval, err := agent.ExecuteRPC(s.ctx, []string{"rm"}, rmSubcommand, storeConfig)
				if err != nil || retval != 0 {
					s.ctx.GetLogger().Error("Error removing obsolete backups: %s", err)
					goto close
				} else {
					s.ctx.GetLogger().Info("Retention purge succeeded")
				}
			}

		close:
			repo.Close()
			store.Close(s.ctx)
		}
	}
}
