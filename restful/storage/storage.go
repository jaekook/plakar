package storage

import (
	"context"
	"io"

	"github.com/PlakarKorp/plakar/restful/models"
)

// Storage interface defines all storage operations
type Storage interface {
	// Repository operations
	CreateRepository(ctx context.Context, req *models.CreateRepositoryRequest) (string, error)
	GetRepositoryInfo(ctx context.Context) (*models.RepositoryInfo, error)
	GetRepositoryStates(ctx context.Context) ([]string, error)
	GetRepositoryState(ctx context.Context, stateID string) ([]byte, error)
	RunMaintenance(ctx context.Context, req *models.MaintenanceRequest) (*models.MaintenanceResponse, error)
	PruneRepository(ctx context.Context, req *models.PruneRequest) (*models.PruneResponse, error)
	SyncRepository(ctx context.Context, req *models.SyncRequest) (*models.SyncResponse, error)

	// Snapshot operations
	ListSnapshots(ctx context.Context, req *models.ListSnapshotsRequest) ([]models.SnapshotHeader, int, error)
	CreateSnapshot(ctx context.Context, req *models.CreateSnapshotRequest) (*models.CreateSnapshotResponse, error)
	GetSnapshotHeader(ctx context.Context, snapshotID string) (*models.SnapshotHeader, error)
	RestoreSnapshot(ctx context.Context, req *models.RestoreSnapshotRequest) (*models.RestoreSnapshotResponse, error)
	CheckSnapshot(ctx context.Context, req *models.CheckSnapshotRequest) (*models.CheckSnapshotResponse, error)
	DiffSnapshots(ctx context.Context, req *models.DiffSnapshotsRequest) (*models.DiffSnapshotsResponse, error)
	MountSnapshot(ctx context.Context, req *models.MountSnapshotRequest) (*models.MountSnapshotResponse, error)
	UnmountSnapshot(ctx context.Context, req *models.UnmountSnapshotRequest) error
	RemoveSnapshots(ctx context.Context, req *models.RemoveSnapshotsRequest) (*models.RemoveSnapshotsResponse, error)

	// VFS operations
	BrowseVFS(ctx context.Context, snapshotID, path string) (*models.VFSEntry, error)
	ListVFSChildren(ctx context.Context, req *models.ListVFSChildrenRequest) (*models.ItemsPageWrapper, error)
	SearchVFS(ctx context.Context, req *models.SearchVFSRequest) (*models.ItemsPageWrapper, error)
	GetVFSChunks(ctx context.Context, snapshotID, path string) (interface{}, error)
	GetVFSErrors(ctx context.Context, snapshotID, path string) (interface{}, error)
	CreateDownloadPackage(ctx context.Context, req *models.CreateDownloadPackageRequest) (*models.CreateDownloadPackageResponse, error)
	GetSignedDownloadURL(ctx context.Context, downloadID string) (string, error)

	// File operations
	ReadFile(ctx context.Context, req *models.ReadFileRequest) ([]byte, string, error)
	CreateSignedURL(ctx context.Context, req *models.CreateSignedURLRequest) (*models.CreateSignedURLResponse, error)
	GetFileContent(ctx context.Context, req *models.GetFileContentRequest) ([]byte, string, error)
	GetFileDigest(ctx context.Context, req *models.GetFileDigestRequest) (*models.GetFileDigestResponse, error)

	// Search operations
	LocatePathname(ctx context.Context, req *models.LocatePathnameRequest) ([]models.TimelineLocation, int, error)
	GetImporterTypes(ctx context.Context) ([]string, error)
	LocateFiles(ctx context.Context, req *models.LocateFilesRequest) (*models.LocateFilesResponse, error)

	// Authentication operations
	LoginGitHub(ctx context.Context, req *models.LoginGitHubRequest) (*models.LoginResponse, error)
	LoginEmail(ctx context.Context, req *models.LoginEmailRequest) (*models.LoginResponse, error)
	Logout(ctx context.Context) error

	// Integration operations
	InstallIntegration(ctx context.Context, req *models.InstallIntegrationRequest) error
	UninstallIntegration(ctx context.Context, integrationID string) error

	// Scheduler operations
	StartScheduler(ctx context.Context) error
	StopScheduler(ctx context.Context) error
	GetSchedulerStatus(ctx context.Context) (*models.SchedulerStatus, error)

	// System operations
	GetAPIInfo(ctx context.Context) (*models.APIInfo, error)
	ProxyRequest(ctx context.Context, method, path string, body io.Reader) (*models.ProxyResponse, error)
}