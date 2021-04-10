package bbs

import (
	"fmt"
	"log"
	"strings"
	"time"
)

// UserRecord mapping to `userec` in most system, it records uesr's
// basical data
type UserRecord interface {
	// UserID return user's identification string, and it is userid in
	// mostly bbs system
	UserID() string
	// HashedPassword return user hashed password, it only for debug,
	// If you want to check is user password correct, please use
	// VerifyPassword insteaded.
	HashedPassword() string
	// VerifyPassword will check user's password is OK. it will return null
	// when OK and error when there are something wrong
	VerifyPassword(password string) error
	// Nickname return a string for user's nickname, this string may change
	// depend on user's mood, return empty string if this bbs system do not support
	Nickname() string
	// RealName return a string for user's real name, this string may not be changed
	// return empty string if this bbs system do not support
	RealName() string
	// NumLoginDays return how many days this have been login since account created.
	NumLoginDays() int
	// NumPosts return how many posts this user has posted.
	NumPosts() int
	// Money return the money this user have.
	Money() int
	// LastLogin return last login time of user
	LastLogin() time.Time
	// LastHost return last login host of user, it is IPv4 address usually, but it
	// could be domain name or IPv6 address.
	LastHost() string
}

// BadPostUserRecord return UserRecord interface which support NumBadPosts
type BadPostUserRecord interface {
	// NumBadPosts return how many bad post this use have
	NumBadPosts() int
}

// LastCountryUserRecord return UserRecord interface which support LastCountry
type LastCountryUserRecord interface {
	// LastLoginCountry will return the country with this user's last login IP
	LastLoginCountry() string
}

// MailboxUserRecord return UserRecord interface which support MailboxDescription
type MailboxUserRecord interface {
	// MailboxDescription will return the mailbox description with this user
	MailboxDescription() string
}

type FavoriteType int

const (
	FavoriteTypeBoard  FavoriteType = iota // 0
	FavoriteTypeFolder                     // 1
	FavoriteTypeLine                       // 2

)

type FavoriteRecord interface {
	Title() string
	Type() FavoriteType
	BoardID() string

	// Records is FavoriteTypeFolder only.
	Records() []FavoriteRecord
}

type BoardRecord interface {
	BoardID() string

	Title() string

	IsClass() bool
	// ClassID should return the class id to which this board/class belongs.
	ClassID() string

	BM() []string
}

type ArticleRecord interface {
	Filename() string
	Modified() time.Time
	Recommend() int
	Date() string
	Title() string
	Money() int
	Owner() string
}

// DB is whole bbs filesystem, including where file store,
// how to connect to local cache ( system V shared memory or etc.)
// how to parse or store it's data to bianry
type DB struct {
	connector Connector
}

// Driver should implement Connector interface
type Connector interface {
	// Open provides the driver parameter settings, such as BBSHome parameter and SHM parameters.
	Open(dataSourceName string) error
	// GetUserRecordsPath should return user records file path, eg: BBSHome/.PASSWDS
	GetUserRecordsPath() (string, error)
	// ReadUserRecordsFile should return UserRecord list in the file called name
	ReadUserRecordsFile(name string) ([]UserRecord, error)
	// GetUserFavoriteRecordsPath should return the user favorite records file path
	// for specific user, eg: BBSHOME/home/{{u}}/{{userID}}/.fav
	GetUserFavoriteRecordsPath(userID string) (string, error)
	// ReadUserFavoriteRecordsFile should return FavoriteRecord list in the file called name
	ReadUserFavoriteRecordsFile(name string) ([]FavoriteRecord, error)
	// GetBoardRecordsPath should return the board headers file path, eg: BBSHome/.BRD
	GetBoardRecordsPath() (string, error)
	// ReadBoardRecordsFile shoule return BoardRecord list in file, name is the file name
	ReadBoardRecordsFile(name string) ([]BoardRecord, error)
	// GetBoardArticleRecordsPath should return the article records file path, boardID is the board id,
	// eg: BBSHome/boards/{{b}}/{{boardID}}/.DIR
	GetBoardArticleRecordsPath(boardID string) (string, error)
	// GetBoardArticleRecordsPath should return the treasure records file path, boardID is the board id,
	// eg: BBSHome/man/boards/{{b}}/{{boardID}}/{{treasureID}}/.DIR
	GetBoardTreasureRecordsPath(boardID string, treasureID []string) (string, error)
	// ReadArticleRecordsFile returns ArticleRecord list in file, name is the file name
	ReadArticleRecordsFile(name string) ([]ArticleRecord, error)
	// GetBoardArticleFilePath return file path for specific boardID and filename
	GetBoardArticleFilePath(boardID string, filename string) (string, error)
	// GetBoardTreasureFilePath return file path for specific boardID, treasureID and filename
	GetBoardTreasureFilePath(boardID string, treasureID []string, name string) (string, error)
	// ReadBoardArticleFile should returns raw file of specific file name
	ReadBoardArticleFile(name string) ([]byte, error)
}

// Driver which implement WriteBoardConnector supports modify board record file.
type WriteBoardConnector interface {

	// NewBoardRecord return BoardRecord object in this driver with arugments
	NewBoardRecord(args map[string]interface{}) (BoardRecord, error)

	// AddBoardRecordFileRecord given record file name and new record, should append
	// file record in that file.
	AddBoardRecordFileRecord(name string, brd BoardRecord) error

	// UpdateBoardRecordFileRecord update boardRecord brd on index in record file,
	// index is start with 0
	UpdateBoardRecordFileRecord(name string, index uint, brd BoardRecord) error

	// ReadBoardRecordFileRecord return boardRecord brd on index in record file.
	ReadBoardRecordFileRecord(name string, index uint) (BoardRecord, error)

	// RemoveBoardRecordFileRecord remove boardRecord brd on index in record file.
	RemoveBoardRecordFileRecord(name string, index uint) error
}

// UserArticleConnector is a connector for bbs who support cached user article records
type UserArticleConnector interface {

	// GetUserArticleRecordsPath should return the file path which user article record stores.
	GetUserArticleRecordsPath(userID string) (string, error)

	// ReadUserArticleRecordFile should return the article record in file.
	ReadUserArticleRecordFile(name string) ([]UserArticleRecord, error)

	// WriteUserArticleRecordFile write user article records into file.
	WriteUserArticleRecordFile(name string, records []UserArticleRecord) error

	// AppendUserArticleRecordFile append user article records into file.
	AppendUserArticleRecordFile(name string, record UserArticleRecord) error
}

var drivers = make(map[string]Connector)

func Register(drivername string, connector Connector) {
	// TODO: Mutex
	drivers[drivername] = connector
}

// Open opan a
func Open(drivername string, dataSourceName string) (*DB, error) {

	c, ok := drivers[drivername]
	if !ok {
		return nil, fmt.Errorf("bbs: drivername: %v not found", drivername)
	}

	err := c.Open(dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("bbs: drivername: %v open error: %v", drivername, err)
	}

	return &DB{
		connector: c,
	}, nil
}

// ReadUserRecords returns the UserRecords
func (db *DB) ReadUserRecords() ([]UserRecord, error) {

	path, err := db.connector.GetUserRecordsPath()
	if err != nil {
		log.Println("bbs: open file error:", err)
		return nil, err
	}
	log.Println("path:", path)

	userRecs, err := db.connector.ReadUserRecordsFile(path)
	if err != nil {
		log.Println("bbs: get user rec error:", err)
		return nil, err
	}
	return userRecs, nil
}

// ReadUserFavoriteRecords returns the FavoriteRecord for specific userID
func (db *DB) ReadUserFavoriteRecords(userID string) ([]FavoriteRecord, error) {

	path, err := db.connector.GetUserFavoriteRecordsPath(userID)
	if err != nil {
		log.Println("bbs: get user favorite records path error:", err)
		return nil, err
	}
	log.Println("path:", path)

	recs, err := db.connector.ReadUserFavoriteRecordsFile(path)
	if err != nil {
		log.Println("bbs: read user favorite records error:", err)
		return nil, err
	}
	return recs, nil

}

// ReadBoardRecords returns the UserRecords
func (db *DB) ReadBoardRecords() ([]BoardRecord, error) {

	path, err := db.connector.GetBoardRecordsPath()
	if err != nil {
		log.Println("bbs: open file error:", err)
		return nil, err
	}
	log.Println("path:", path)

	recs, err := db.connector.ReadBoardRecordsFile(path)
	if err != nil {
		log.Println("bbs: get user rec error:", err)
		return nil, err
	}
	return recs, nil
}

func (db *DB) ReadBoardArticleRecordsFile(boardID string) ([]ArticleRecord, error) {

	path, err := db.connector.GetBoardArticleRecordsPath(boardID)
	if err != nil {
		log.Println("bbs: open file error:", err)
		return nil, err
	}
	log.Println("path:", path)

	recs, err := db.connector.ReadArticleRecordsFile(path)
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			return []ArticleRecord{}, nil
		}
		log.Println("bbs: ReadArticleRecordsFile error:", err)
		return nil, err
	}
	return recs, nil

}

func (db *DB) ReadBoardTreasureRecordsFile(boardID string, treasureID []string) ([]ArticleRecord, error) {

	path, err := db.connector.GetBoardTreasureRecordsPath(boardID, treasureID)
	if err != nil {
		log.Println("bbs: open file error:", err)
		return nil, err
	}
	log.Println("path:", path)

	recs, err := db.connector.ReadArticleRecordsFile(path)
	if err != nil {
		log.Println("bbs: get user rec error:", err)
		return nil, err
	}
	return recs, nil
}

func (db *DB) ReadBoardArticleFile(boardID string, filename string) ([]byte, error) {

	path, err := db.connector.GetBoardArticleFilePath(boardID, filename)
	if err != nil {
		log.Println("bbs: open file error:", err)
		return nil, err
	}
	log.Println("path:", path)

	recs, err := db.connector.ReadBoardArticleFile(path)
	if err != nil {
		log.Println("bbs: get user rec error:", err)
		return nil, err
	}
	return recs, nil
}

func (db *DB) ReadBoardTreasureFile(boardID string, treasuresID []string, filename string) ([]byte, error) {

	path, err := db.connector.GetBoardTreasureFilePath(boardID, treasuresID, filename)
	if err != nil {
		log.Println("bbs: open file error:", err)
		return nil, err
	}
	log.Println("path:", path)

	recs, err := db.connector.ReadBoardArticleFile(path)
	if err != nil {
		log.Println("bbs: get user rec error:", err)
		return nil, err
	}
	return recs, nil
}

func (db *DB) NewBoardRecord(args map[string]interface{}) (BoardRecord, error) {
	return db.connector.(WriteBoardConnector).NewBoardRecord(args)
}

func (db *DB) AddBoardRecord(brd BoardRecord) error {

	path, err := db.connector.GetBoardRecordsPath()
	if err != nil {
		log.Println("bbs: open file error:", err)
		return err
	}
	log.Println("path:", path)

	err = db.connector.(WriteBoardConnector).AddBoardRecordFileRecord(path, brd)
	if err != nil {
		log.Println("bbs: AddBoardRecordFileRecord error:", err)
		return err
	}
	return nil
}

// UpdateBoardRecordFileRecord update boardRecord brd on index in record file,
// index is start with 0
func (db *DB) UpdateBoardRecord(index uint, brd *BoardRecord) error {
	return fmt.Errorf("not implement")
}

// ReadBoardRecordFileRecord return boardRecord brd on index in record file.
func (db *DB) ReadBoardRecord(index uint) (*BoardRecord, error) {
	return nil, fmt.Errorf("not implement")
}

// RemoveBoardRecordFileRecord remove boardRecord brd on index in record file.
func (db *DB) RemoveBoardRecord(index uint) error {
	return fmt.Errorf("not implement")
}

// GetUserArticleRecordFile returns aritcle file which user posted.
func (db *DB) GetUserArticleRecordFile(userID string) ([]UserArticleRecord, error) {

	recs := []UserArticleRecord{}
	uac, ok := db.connector.(UserArticleConnector)
	if ok {

		path, err := uac.GetUserArticleRecordsPath(userID)
		if err != nil {
			log.Println("bbs: open file error:", err)
			return nil, err
		}
		log.Println("path:", path)

		recs, err = uac.ReadUserArticleRecordFile(path)
		if err != nil {
			log.Println("bbs: ReadUserArticleRecordFile error:", err)
			return nil, err
		}
		if len(recs) != 0 {
			return recs, nil
		}

	}

	boardRecords, err := db.ReadBoardRecords()
	if err != nil {
		log.Println("bbs: ReadBoardRecords error:", err)
		return nil, err
	}

	shouldSkip := func(boardID string) bool {
		if boardID == "ALLPOST" {
			return true
		}
		return false
	}

	for _, r := range boardRecords {
		if shouldSkip(r.BoardID()) {
			continue
		}

		ars, err := db.ReadBoardArticleRecordsFile(r.BoardID())
		if err != nil {
			log.Println("bbs: ReadBoardArticleRecordsFile error:", err)
			return nil, err
		}
		for _, ar := range ars {
			if ar.Owner() == userID {
				log.Println("board: ", r.BoardID(), len(recs))
				r := userArticleRecord{
					"board_id":   r.BoardID(),
					"title":      ar.Title(),
					"owner":      ar.Owner(),
					"article_id": ar.Filename(),
				}
				recs = append(recs, r)
			}
		}
	}

	return recs, nil
}
