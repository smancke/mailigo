package mail

import (
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/mattn/go-sqlite3"
	"github.com/smancke/mailigo/logging"
	"io/ioutil"
	"os"
	"path"
)

// The Sqlite filename within the library directory
var DBFilename = ".library.db"
var WriteTestFilename = ".galleryWriteTest"

type DBMailing struct {
	ID           string `sql:"type:varchar(50)"gorm:"primary_key"`
	TemplateName string `sql:"type:varchar(50)"`
	GlobalData   string `sql:"type:text"`
}

type ProcessingResult struct {
	MailingID string
	To        string
	HasError  bool
	Error     string
}

type Store struct {
	db  *gorm.DB
	dir string
}

func NewStore() *Store {
	return &Store{}
}

// Opens the store denoted by the given directory.
// If the directory does not exist, it will be created.
func (store *Store) Open(directoryPath string) error {
	if err := ensureWriteableDirectory(directoryPath); err != nil {
		return err
	}
	dbFilename := path.Join(directoryPath, DBFilename)
	if err := store.openDB(dbFilename); err != nil {
		return err
	}
	store.dir = directoryPath
	return nil
}

func (store *Store) openDB(filename string) error {
	logging.Logger.Infof("opening sqlite3 db: %v", filename)
	gormdb, err := gorm.Open("sqlite3", filename)
	if err == nil {
		if err := gormdb.DB().Ping(); err != nil {
			logging.Logger.Infof("error pinging database: %v\n", err)
		} else {
			logging.Logger.Infof("can ping database")
		}

		//gormdb.LogMode(true)
		gormdb.DB().SetMaxIdleConns(2)
		gormdb.DB().SetMaxOpenConns(5)
		gormdb.SingularTable(true)

		if err := gormdb.AutoMigrate(&DBMailing{}, &ProcessingResult{}).Error; err != nil {
			logging.Logger.Infof("error in schema migration: %v", err)
			return err
		} else {
			logging.Logger.Infof("ensured db schema")
		}
	} else {
		logging.Logger.Infof("error opening sqlite3 db %v: %v\n", filename, err)
	}
	store.db = gormdb
	return err
}

func (store *Store) CreateMailing(mailing DBMailing) error {
	return store.db.Create(mailing).Error
}

func (store *Store) Close() (err error) {
	logging.Logger.Infof("closing sqlite3 db")
	return store.db.Close()
}
func ensureWriteableDirectory(dir string) error {
	dirInfo, err := os.Stat(dir)
	if os.IsNotExist(err) {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		dirInfo, err = os.Stat(dir)
	}

	if err != nil || !dirInfo.IsDir() {
		return fmt.Errorf("not a directory %v", dir)
	}

	writeTest := path.Join(dir, WriteTestFilename)
	if err := ioutil.WriteFile(writeTest, []byte("writeTest"), 0644); err != nil {
		return err
	}
	if err := os.Remove(writeTest); err != nil {
		return err
	}
	return nil
}
