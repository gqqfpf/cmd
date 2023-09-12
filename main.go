package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/snapshot"

	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"github.com/gqq/backupcmd/pkg/file"

	"os"
	"path/filepath"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"time"
)

func loggedError(log logr.Logger, err error, message string) error {
	log.Error(err, message)
	return fmt.Errorf("%s : %s\n", message, err)
}

func main() {
	var (
		backupTempDir          string
		etcdURL                string
		etcdDialTimeoutSeconds int64
		timeoutSeconds         int64
		bucketname             string
		objectname             string
	)

	flag.StringVar(&backupTempDir, "backup-temp-dir", os.TempDir(), "the directory to temp place backups before they are uploaded to their destination")
	flag.StringVar(&etcdURL, "etcd-url", "http://localhost2379", "url for etcd")
	flag.Int64Var(&etcdDialTimeoutSeconds, "etcd-dial-timeout-seconds", 5, "Timeout,in seconds for dialing the Etcd API")
	flag.Int64Var(&timeoutSeconds, "timeout-seconds", 60, "Timeout, in seconds, of the whole restore operation.")
	flag.StringVar(&bucketname, "bucketname", "gqq", "Remote s3 bucket name")
	flag.StringVar(&objectname, "objectname", "snapshot.db", "Remote S3 bucket object name")
	flag.Parse()
	newlogger := zap.NewRaw(zap.UseDevMode(true))
	ctrl.SetLogger(zapr.NewLogger(newlogger))
	logger := ctrl.Log.WithName("backup-agent")

	fmt.Printf("etcdURL is %s\n", etcdURL)
	fmt.Printf("etcdDialTimeoutSeconds is %d\n", etcdDialTimeoutSeconds)

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*time.Duration(timeoutSeconds))
	defer cancelFunc()

	logger.Info("Connecting to Etcd and getting snapshot")
	localpath := filepath.Join(backupTempDir, "snapshot.db")

	etcdClientv3 := snapshot.NewV3(newlogger.Named("etcd-client"))
	err := etcdClientv3.Save(
		ctx,
		clientv3.Config{
			Endpoints:   []string{etcdURL},
			DialTimeout: time.Second * time.Duration(etcdDialTimeoutSeconds),
		},
		localpath,
	)
	if err != nil {
		panic(loggedError(logger, err, "failed to get etcd snapshot"))
	}
	endpoint := "http://192.168.251.176:30000"
	assessKeyID := "admin"
	secretAccessKeyID := "edoc2@edoc2"
	uploader := file.News3Uploader(endpoint, assessKeyID, secretAccessKeyID)
	logger.Info("Uplaoding snapshot")

	uploadsize, err := uploader.Upload(ctx, localpath, bucketname, objectname)
	if err != nil {
		panic(loggedError(logger, err, "failed to upload backup"))
	}
	logger.WithValues("upload-size", uploadsize).Info("Backup complete")
}
