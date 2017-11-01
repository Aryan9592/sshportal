package main

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/jinzhu/gorm"
)

type SSHKey struct {
	// FIXME: use uuid for ID
	gorm.Model
	Name        string // FIXME: govalidator: min length 3, alphanum
	Type        string
	Length      uint
	Fingerprint string
	PrivKey     string
	PubKey      string
}

type Host struct {
	// FIXME: use uuid for ID
	gorm.Model
	Name        string `gorm:"unique_index"` // FIXME: govalidator: min length 3, alphanum
	Addr        string
	User        string
	Password    string
	Fingerprint string
	PrivKey     *SSHKey
	Groups      []Group `gorm:"many2many:host_groups;"`
}

type Group struct {
	// FIXME: use uuid for ID
	gorm.Model
	Name string `gorm:"unique_index"` // FIXME: govalidator: min length 3, alphanum
}

type User struct {
	// FIXME: use uuid for ID
	gorm.Model
	Name    string `gorm:"unique_index"` // FIXME: govalidator: min length 3, alphanum
	SSHKeys []SSHKey
	Groups  []Group `gorm:"many2many:user_groups;"`
}

func dbInit(db *gorm.DB) error {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&SSHKey{})
	db.AutoMigrate(&Host{})
	db.AutoMigrate(&Group{})
	return nil
}

func dbDemo(db *gorm.DB) error {
	var host1, host2, host3 Host
	db.FirstOrCreate(&host1, &Host{Name: "sdf", Addr: "sdf.org:22", User: "new"})
	db.FirstOrCreate(&host2, &Host{Name: "whoami", Addr: "whoami.filippo.io:22", User: "test"})
	db.FirstOrCreate(&host3, &Host{Name: "ssh-chat", Addr: "chat.shazow.net:22", User: "test", Fingerprint: "MD5:e5:d5:d1:75:90:38:42:f6:c7:03:d7:d0:56:7d:6a:db"})
	return nil
}

func RemoteHostFromSession(s ssh.Session, db *gorm.DB) (*Host, error) {
	var host Host
	db.Where("name = ?", s.User()).Find(&host)
	if host.Name == "" {
		// FIXME: add available hosts
		return nil, fmt.Errorf("No such target: %q", s.User())
	}
	return &host, nil
}

func (host *Host) URL() string {
	return fmt.Sprintf("%s@%s", host.User, host.Addr)
}

func NewHostFromURL(rawurl string) (*Host, error) {
	if !strings.Contains(rawurl, "://") {
		rawurl = "ssh://" + rawurl
	}
	u, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}
	host := Host{Addr: u.Host}
	if u.User != nil {
		password, _ := u.User.Password()
		host.Password = password
		host.User = u.User.Username()
	}
	return &host, nil
}

func (host *Host) Hostname() string {
	return strings.Split(host.Addr, ":")[0]
}

func FindHostByIdOrName(db *gorm.DB, query string) (*Host, error) {
	var host Host
	if err := db.Where("id = ?", query).Or("name = ?", query).First(&host).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

func FindHostsByIdOrName(db *gorm.DB, queries []string) ([]*Host, error) {
	var hosts []*Host
	for _, query := range queries {
		host, err := FindHostByIdOrName(db, query)
		if err != nil {
			return nil, err
		}
		hosts = append(hosts, host)
	}
	return hosts, nil
}

func FindKeyByIdOrName(db *gorm.DB, query string) (*SSHKey, error) {
	var key SSHKey
	if err := db.Where("id = ?", query).Or("name = ?", query).First(&key).Error; err != nil {
		return nil, err
	}
	return &key, nil
}

func FindKeysByIdOrName(db *gorm.DB, queries []string) ([]*SSHKey, error) {
	var keys []*SSHKey
	for _, query := range queries {
		key, err := FindKeyByIdOrName(db, query)
		if err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, nil
}
