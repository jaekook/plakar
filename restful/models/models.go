package models

import (
	"time"
)

// Generic wrapper types
type ItemWrapper struct {
	Item interface{} `json:"item"`
}

type ItemsWrapper struct {
	Total int           `json:"total"`
	Items []interface{} `json:"items"`
}

type ItemsPageWrapper struct {
	HasNext bool          `json:"has_next"`
	Items   []interface{} `json:"items"`
}

// Repository models

type CreateRepositoryRequest struct {
	Location      string `json:"location" validate:"required"`
	Passphrase    string `json:"passphrase"`
	Hashing       string `json:"hashing"`
	Plaintext     bool   `json:"plaintext"`
	NoCompression bool   `json:"no_compression"`
}

type CreateRepositoryResponse struct {
	RepositoryID string `json:"repository_id"`
	Location     string `json:"location"`
}

type RepositoryInfo struct {
	Location      string                  `json:"location"`
	Snapshots     RepositoryInfoSnapshots `json:"snapshots"`
	Configuration RepositoryConfiguration `json:"configuration"`
	OS            string                  `json:"os"`
	Arch          string                  `json:"arch"`
}

type RepositoryInfoSnapshots struct {
	Total           int     `json:"total"`
	StorageSize     int64   `json:"storage_size"`
	LogicalSize     int64   `json:"logical_size"`
	Efficiency      float64 `json:"efficiency"`
	SnapshotsPerDay []int   `json:"snapshots_per_day"`
}

type RepositoryConfiguration struct {
	RepositoryID string    `json:"repository_id"`
	Version      string    `json:"version"`
	Timestamp    time.Time `json:"timestamp"`
	Encryption   *struct{} `json:"encryption,omitempty"`
}

type ListSnapshotsRequest struct {
	Offset   uint32     `json:"offset"`
	Limit    uint32     `json:"limit"`
	Importer string     `json:"importer"`
	Since    *time.Time `json:"since"`
	Sort     string     `json:"sort"`
}

type LocatePathnameRequest struct {
	Resource          string `json:"resource"`
	ImporterType      string `json:"importer_type"`
	ImporterOrigin    string `json:"importer_origin"`
	ImporterDirectory string `json:"importer_directory"`
	Offset            uint32 `json:"offset"`
	Limit             uint32 `json:"limit"`
	Sort              string `json:"sort"`
}

type ImporterType struct {
	Name string `json:"name"`
}

type MaintenanceRequest struct {
	Operations []string `json:"operations"`
}

type MaintenanceResponse struct {
	OperationsPerformed []string `json:"operations_performed"`
	Duration            string   `json:"duration"`
	CleanedObjects      int      `json:"cleaned_objects"`
}

type PruneRequest struct {
	Policy  string                 `json:"policy"`
	Filters map[string]interface{} `json:"filters"`
	Apply   bool                   `json:"apply"`
}

type PruneResponse struct {
	ReclaimedSpace int64 `json:"reclaimed_space"`
	RemovedObjects int   `json:"removed_objects"`
	DryRun         bool  `json:"dry_run"`
}

type SyncRequest struct {
	TargetRepository string                 `json:"target_repository" validate:"required"`
	Direction        string                 `json:"direction" validate:"required,oneof=to from with"`
	SnapshotIDs      []string               `json:"snapshot_ids"`
	Filters          map[string]interface{} `json:"filters"`
}

type SyncResponse struct {
	SyncedSnapshots int64  `json:"synced_snapshots"`
	TransferredSize int64  `json:"transferred_size"`
	Direction       string `json:"direction"`
}

// Snapshot models

type CreateSnapshotRequest struct {
	Source           string            `json:"source" validate:"required"`
	Tags             []string          `json:"tags"`
	Excludes         []string          `json:"excludes"`
	Concurrency      int               `json:"concurrency"`
	DryRun           bool              `json:"dry_run"`
	Check            bool              `json:"check"`
	ForcedTimestamp  *time.Time        `json:"forced_timestamp"`
	Options          map[string]string `json:"options"`
}

type CreateSnapshotResponse struct {
	SnapshotID string          `json:"snapshot_id"`
	Timestamp  time.Time       `json:"timestamp"`
	Summary    SnapshotSummary `json:"summary"`
}

type SnapshotHeader struct {
	ID        string            `json:"id"`
	Timestamp time.Time         `json:"timestamp"`
	Duration  int64             `json:"duration"`
	Summary   SnapshotSummary   `json:"summary"`
	Sources   []SnapshotSource  `json:"sources"`
	Tags      []string          `json:"tags"`
}

type SnapshotSummary struct {
	Files       int   `json:"files"`
	Directories int   `json:"directories"`
	Size        int64 `json:"size"`
}

type SnapshotSource struct {
	Importer SnapshotImporter `json:"importer"`
}

type SnapshotImporter struct {
	Type      string `json:"type"`
	Origin    string `json:"origin"`
	Directory string `json:"directory"`
}

type RestoreSnapshotRequest struct {
	SnapshotID        string                 `json:"snapshot_id"`
	Destination       string                 `json:"destination" validate:"required"`
	Paths             []string               `json:"paths"`
	Concurrency       int                    `json:"concurrency"`
	SkipPermissions   bool                   `json:"skip_permissions"`
	Filters           map[string]interface{} `json:"filters"`
}

type RestoreSnapshotResponse struct {
	RestoredFiles int    `json:"restored_files"`
	RestoredSize  int64  `json:"restored_size"`
	Destination   string `json:"destination"`
}

type CheckSnapshotRequest struct {
	SnapshotID  string   `json:"snapshot_id"`
	Paths       []string `json:"paths"`
	Fast        bool     `json:"fast"`
	NoVerify    bool     `json:"no_verify"`
	Concurrency int      `json:"concurrency"`
}

type CheckSnapshotResponse struct {
	Status            string       `json:"status"`
	CheckedFiles      int          `json:"checked_files"`
	Errors            []CheckError `json:"errors"`
	SignatureVerified bool         `json:"signature_verified"`
}

type CheckError struct {
	Path  string `json:"path"`
	Error string `json:"error"`
}

type DiffSnapshotsRequest struct {
	SnapshotID       string `json:"snapshot_id"`
	TargetSnapshotID string `json:"target_snapshot_id"`
	Path             string `json:"path"`
	Recursive        bool   `json:"recursive"`
	Highlight        bool   `json:"highlight"`
}

type DiffSnapshotsResponse struct {
	Changes []DiffChange `json:"changes"`
}

type DiffChange struct {
	Path string `json:"path"`
	Type string `json:"type"`
	Diff string `json:"diff"`
}

type MountSnapshotRequest struct {
	SnapshotID string `json:"snapshot_id"`
	Mountpoint string `json:"mountpoint" validate:"required"`
}

type MountSnapshotResponse struct {
	Mountpoint string `json:"mountpoint"`
	SnapshotID string `json:"snapshot_id"`
}

type UnmountSnapshotRequest struct {
	SnapshotID string `json:"snapshot_id"`
	Mountpoint string `json:"mountpoint" validate:"required"`
}

type RemoveSnapshotsRequest struct {
	SnapshotIDs []string               `json:"snapshot_ids"`
	Filters     map[string]interface{} `json:"filters"`
	Apply       bool                   `json:"apply"`
}

type RemoveSnapshotsResponse struct {
	RemovedSnapshots []string `json:"removed_snapshots"`
	DryRun           bool     `json:"dry_run"`
}

// VFS models

type VFSEntry struct {
	Name     string    `json:"name"`
	Path     string    `json:"path"`
	Type     string    `json:"type"`
	Size     int64     `json:"size"`
	Mode     int       `json:"mode"`
	MTime    time.Time `json:"mtime"`
	UID      int       `json:"uid"`
	GID      int       `json:"gid"`
	Checksum string    `json:"checksum,omitempty"`
	Target   string    `json:"target,omitempty"`
}

type ListVFSChildrenRequest struct {
	SnapshotID string `json:"snapshot_id"`
	Path       string `json:"path"`
	Offset     uint32 `json:"offset"`
	Limit      uint32 `json:"limit"`
	Sort       string `json:"sort"`
}

type SearchVFSRequest struct {
	SnapshotID string `json:"snapshot_id"`
	Path       string `json:"path"`
	Pattern    string `json:"pattern"`
	Offset     uint32 `json:"offset"`
	Limit      uint32 `json:"limit"`
}

type CreateDownloadPackageRequest struct {
	SnapshotID string   `json:"snapshot_id"`
	Path       string   `json:"path"`
	Files      []string `json:"files" validate:"required"`
	Format     string   `json:"format"`
	Rebase     bool     `json:"rebase"`
}

type CreateDownloadPackageResponse struct {
	DownloadID  string `json:"download_id"`
	DownloadURL string `json:"download_url"`
}

type SignedURLResponse struct {
	URL string `json:"url"`
}

// File operation models

type ReadFileRequest struct {
	SnapshotID string `json:"snapshot_id"`
	Path       string `json:"path"`
	Download   bool   `json:"download"`
	Render     string `json:"render"`
}

type CreateSignedURLRequest struct {
	SnapshotID string `json:"snapshot_id"`
	Path       string `json:"path"`
}

type CreateSignedURLResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

type GetFileContentRequest struct {
	SnapshotID string `json:"snapshot_id"`
	Path       string `json:"path"`
	Decompress bool   `json:"decompress"`
	Highlight  bool   `json:"highlight"`
}

type GetFileDigestRequest struct {
	SnapshotID string `json:"snapshot_id"`
	Path       string `json:"path"`
	Algorithm  string `json:"algorithm"`
}

type GetFileDigestResponse struct {
	Path      string `json:"path"`
	Algorithm string `json:"algorithm"`
	Digest    string `json:"digest"`
	Size      int64  `json:"size"`
}

// Search models

type TimelineLocation struct {
	Snapshot SnapshotHeader `json:"snapshot"`
	VFSEntry VFSEntry       `json:"vfs_entry"`
}

type LocateFilesRequest struct {
	Patterns []string `json:"patterns"`
	Snapshot string   `json:"snapshot"`
	Filters  string   `json:"filters"`
	Limit    uint32   `json:"limit"`
}

type LocateFilesResponse struct {
	Matches []FileMatch `json:"matches"`
}

type FileMatch struct {
	SnapshotID string   `json:"snapshot_id"`
	Path       string   `json:"path"`
	Entry      VFSEntry `json:"entry"`
}

// Authentication models

type LoginGitHubRequest struct {
	Redirect string `json:"redirect"`
}

type LoginEmailRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Redirect string `json:"redirect"`
}

type LoginResponse struct {
	URL string `json:"URL"`
}

// Integration models

type InstallIntegrationRequest struct {
	Type   string                 `json:"type" validate:"required"`
	Config map[string]interface{} `json:"config"`
}

// Scheduler models

type SchedulerStatus struct {
	Running    bool       `json:"running"`
	NextRun    *time.Time `json:"next_run"`
	ActiveJobs int        `json:"active_jobs"`
}

// System models

type APIInfo struct {
	RepositoryID  string `json:"repository_id"`
	Authenticated bool   `json:"authenticated"`
	Version       string `json:"version"`
	Browsable     bool   `json:"browsable"`
	DemoMode      bool   `json:"demo_mode"`
}

type ProxyResponse struct {
	StatusCode  int                 `json:"status_code"`
	ContentType string              `json:"content_type"`
	Headers     map[string][]string `json:"headers"`
	Body        []byte              `json:"body"`
}