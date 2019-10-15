package fsm

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotmesh-io/dotmesh/pkg/archiver"
	"github.com/dotmesh-io/dotmesh/pkg/types"
	"github.com/dotmesh-io/dotmesh/pkg/utils"

	log "github.com/sirupsen/logrus"
)

func (f *FsMachine) saveFile(file *types.InputFile) StateFn {
	// create the default paths
	destPath := fmt.Sprintf("%s/%s/%s", utils.Mnt(f.filesystemId), "__default__", file.Filename)

	l := log.WithFields(log.Fields{
		"filename": file.Filename,
		"destPath": destPath,
	})

	directoryPath := destPath[:strings.LastIndex(destPath, "/")]
	err := os.MkdirAll(directoryPath, 0775)
	if err != nil {
		e := types.Event{
			Name: types.EventNameSaveFailed,
			Args: &types.EventArgs{"err": fmt.Errorf("failed to create directory, error: %s", err)},
		}
		l.WithField("directoryPath", directoryPath).WithError(err).Error("[saveFile] Error creating directory")
		file.Response <- &e
		return backoffState
	}
	out, err := os.Create(destPath)
	if err != nil {
		e := types.Event{
			Name: types.EventNameSaveFailed,
			Args: &types.EventArgs{"err": fmt.Errorf("failed to create file, error: %s", err)},
		}
		l.WithError(err).Error("[saveFile] Error creating file")
		file.Response <- &e
		return backoffState
	}

	defer func() {
		err := out.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"file":  destPath,
			}).Error("s3 saveFile: got error while closing output file")
		}
	}()

	bytes, err := io.Copy(out, file.Contents)
	if err != nil {
		e := types.Event{
			Name: types.EventNameSaveFailed,
			Args: &types.EventArgs{"err": fmt.Errorf("cannot to create a file, error: %s", err)},
		}
		l.WithError(err).Error("[saveFile] Error writing file")
		file.Response <- &e
		return backoffState
	}
	response, _ := f.snapshot(&types.Event{Name: "snapshot",
		Args: &types.EventArgs{"metadata": map[string]string{
			"message":      "Uploaded " + file.Filename + " (" + formatBytes(bytes) + ")",
			"author":       file.User,
			"type":         "upload",
			"upload.type":  "S3",
			"upload.file":  file.Filename,
			"upload.bytes": fmt.Sprintf("%d", bytes),
		}}})
	if response.Name != "snapshotted" {
		e := types.Event{
			Name: types.EventNameSaveFailed,
			Args: &types.EventArgs{"err": "file snapshot failed"},
		}
		l.WithFields(log.Fields{
			"responseName": response.Name,
			"responseArgs": fmt.Sprintf("%#v", *(response.Args)),
		}).Error("[saveFile] Error committing")
		file.Response <- &e
		return backoffState
	}

	file.Response <- &types.Event{
		Name: types.EventNameSaveSuccess,
		Args: &types.EventArgs{},
	}

	return activeState
}

func (f *FsMachine) readFile(file *types.OutputFile) StateFn {

	// create the default paths
	sourcePath := fmt.Sprintf("%s/%s/%s", file.SnapshotMountPath, "__default__", file.Filename)

	l := log.WithFields(log.Fields{
		"filename":   file.Filename,
		"sourcePath": sourcePath,
	})

	fi, err := os.Stat(sourcePath)
	if err != nil {
		if os.IsNotExist(err) {
			file.Response <- &types.Event{
				Name: types.EventNameFileNotFound,
				Args: &types.EventArgs{"err": fmt.Errorf("failed to stat %s, error: %s", file.Filename, err)},
			}
		} else {
			file.Response <- &types.Event{
				Name: types.EventNameReadFailed,
				Args: &types.EventArgs{"err": fmt.Errorf("failed to stat %s, error: %s", file.Filename, err)},
			}
		}
		l.WithError(err).Error("[readFile] Error statting")
		return backoffState
	}

	if fi.IsDir() {
		return f.readDirectory(file)
	}

	fileOnDisk, err := os.Open(sourcePath)
	if err != nil {
		file.Response <- &types.Event{
			Name: types.EventNameReadFailed,
			Args: &types.EventArgs{"err": fmt.Errorf("failed to read file, error: %s", err)},
		}
		l.WithError(err).Error("[readFile] Error opening")
		return backoffState
	}
	defer func() {
		err := fileOnDisk.Close()
		if err != nil {
			log.WithFields(log.Fields{
				"error": err,
				"file":  sourcePath,
			}).Error("s3 readFile: got error while closing file")
		}
	}()
	_, err = io.Copy(file.Contents, fileOnDisk)
	if err != nil {
		file.Response <- &types.Event{
			Name: types.EventNameReadFailed,
			Args: &types.EventArgs{"err": fmt.Errorf("cannot stream file, error: %s", err)},
		}
		l.WithError(err).Error("[readFile] Error reading")
		return backoffState
	}

	file.Response <- &types.Event{
		Name: types.EventNameReadSuccess,
		Args: &types.EventArgs{},
	}

	return activeState
}

func (f *FsMachine) readDirectory(file *types.OutputFile) StateFn {

	dirPath := filepath.Join(file.SnapshotMountPath, "__default__", file.Filename)

	l := log.WithFields(log.Fields{
		"filename": file.Filename,
		"dirPath":  dirPath,
	})

	stat, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			file.Response <- types.NewErrorEvent(types.EventNameFileNotFound, fmt.Errorf("failed to stat dir '%s', error: %s ", file.Filename, err))
		} else {
			file.Response <- types.NewErrorEvent(types.EventNameReadFailed, fmt.Errorf("failed to stat dir '%s', error: %s ", file.Filename, err))
		}
		l.WithError(err).Error("[readDirectory] Error statting")
		return backoffState
	}

	if !stat.IsDir() {
		file.Response <- types.NewErrorEvent(types.EventNameReadFailed, fmt.Errorf("path '%s' is not a directory, error: %s ", file.Filename, err))
		l.WithError(err).Error("[readDirectory] It isn't a directory")
		return backoffState
	}

	err = archiver.NewTar().ArchiveToStream(file.Contents, []string{dirPath})
	if err != nil {
		file.Response <- types.NewErrorEvent(types.EventNameReadFailed, fmt.Errorf("path '%s' tar failed, error: %s ", file.Filename, err))
		l.WithError(err).Error("[readDirectory] Cannot create tar stream")
		return backoffState
	}

	file.Response <- &types.Event{
		Name: types.EventNameReadSuccess,
		Args: &types.EventArgs{},
	}

	return activeState

}
