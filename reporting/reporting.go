package reporting

import (
	"fmt"
	"os"
	"time"

	"github.com/PlakarKorp/kloset/logging"
	"github.com/PlakarKorp/kloset/objects"
	"github.com/PlakarKorp/kloset/repository"
	"github.com/PlakarKorp/kloset/snapshot"
	"github.com/PlakarKorp/plakar/appcontext"
	"github.com/PlakarKorp/plakar/services"
)

const PLAKAR_API_URL = "https://api.plakar.io/v1/reporting/reports"

type Emitter interface {
	Emit(report *Report, logger *logging.Logger)
}

type Reporter struct {
	logger  *logging.Logger
	emitter Emitter
	reports chan *Report
	stop    chan any
	done    chan any
}

func ReportingEnabled(ctx *appcontext.AppContext) bool {
	authToken, err := ctx.GetCookies().GetAuthToken()
	if err != nil || authToken == "" {
		return false
	}

	sc := services.NewServiceConnector(ctx, authToken)
	enabled, err := sc.GetServiceStatus("alerting")
	if err != nil || !enabled {
		return false
	}

	return true
}

func NewReporter(ctx *appcontext.AppContext) *Reporter {
	logger := ctx.GetLogger()
	var emitter Emitter

	if ReportingEnabled(ctx) {
		emitter = &NullEmitter{}
	} else {

		url := os.Getenv("PLAKAR_API_URL")
		if url == "" {
			url = PLAKAR_API_URL
		}

		var token string

		token, err := ctx.GetCookies().GetAuthToken()
		if err != nil {
			logger.Warn("cannot get auth token")
		}

		emitter = &HttpEmitter{
			url:   url,
			token: token,
			retry: 3,
		}
	}

	r := &Reporter{
		logger:  logger,
		emitter: emitter,
		reports: make(chan *Report, 100),
		stop:    make(chan any),
		done:    make(chan any),
	}

	go func() {
		var rp *Report
		for {
			select {
			case <-ctx.Done():
				goto done
			case <-r.stop:
				goto done
			case rp = <-r.reports:
				r.Process(rp)
			}
		}
	done:
		close(r.reports)
		// drain remaining reports
		for rp = range r.reports {
			r.Process(rp)
		}
		close(r.done)
	}()

	return r
}

func (reporter *Reporter) Process(report *Report) {
	if report.ignore {
		return
	}
	reporter.emitter.Emit(report, reporter.logger)
}

func (reporter *Reporter) StopAndWait() {
	close(reporter.stop)
	for _ = range reporter.done {
	}
}

func (reporter *Reporter) NewReport() *Report {
	return NewReport(reporter.logger, reporter.reports)
}

func NewReport(logger *logging.Logger, reporter chan *Report) *Report {
	report := &Report{
		logger:   logger,
		reporter: reporter,
	}
	return report
}

func (report *Report) SetIgnore() {
	report.ignore = true
}

func (report *Report) TaskStart(kind string, name string) {
	if report.Task != nil {
		report.logger.Warn("already in a task")
	}
	report.Task = &ReportTask{
		StartTime: time.Now(),
		Type:      kind,
		Name:      name,
	}
}

func (report *Report) WithRepositoryName(name string) {
	if report.Repository != nil {
		report.logger.Warn("already has a repository")
	}
	report.Repository = &ReportRepository{
		Name: name,
	}
}

func (report *Report) WithRepository(repository *repository.Repository) {
	report.repo = repository
	configuration := repository.Configuration()
	report.Repository.Storage = configuration
}

func (report *Report) WithSnapshotID(snapshotId objects.MAC) {
	snap, err := snapshot.Load(report.repo, snapshotId)
	if err != nil {
		report.logger.Warn("failed to load snapshot: %s", err)
		return
	}
	report.WithSnapshot(snap)
	snap.Close()
}

func (report *Report) WithSnapshot(snapshot *snapshot.Snapshot) {
	if report.Snapshot != nil {
		report.logger.Warn("already has a snapshot")
	}
	report.Snapshot = &ReportSnapshot{
		Header: *snapshot.Header,
	}
}

func (report *Report) TaskDone() {
	report.taskEnd(StatusOK, 0, "")
}

func (report *Report) TaskWarning(errorMessage string, args ...interface{}) {
	report.taskEnd(StatusWarning, 0, errorMessage, args...)
}

func (report *Report) TaskFailed(errorCode TaskErrorCode, errorMessage string, args ...interface{}) {
	report.taskEnd(StatusFailed, errorCode, errorMessage, args...)
}

func (report *Report) taskEnd(status TaskStatus, errorCode TaskErrorCode, errorMessage string, args ...interface{}) {
	report.Task.Status = status
	report.Task.ErrorCode = errorCode
	if len(args) == 0 {
		report.Task.ErrorMessage = errorMessage
	} else {
		report.Task.ErrorMessage = fmt.Sprintf(errorMessage, args...)
	}
	report.Task.Duration = time.Since(report.Task.StartTime)
	report.Timestamp = time.Now()
	report.reporter <- report
}
