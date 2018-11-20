package db

import (
	"net/url"
	"strings"
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Config struct {
	Hosts []string
	// AuthSource Specify the database name associated with the user’s credentials.
	// authSource defaults to the database specified in the connection string.
	AuthSource string
	UserName   string
	Password   string
	Database   string
	Timeout    time.Duration // second
}

type Mdb struct {
	*Config
	*mgo.Session
}

func NewMdb(c *Config) (*Mdb, error) {
	m := &Mdb{
		Config: c,
	}
	return m, m.connect()
}

func (m *Mdb) connect() error {
	// connectionString: [mongodb://][user:pass@]host1[:port1][,host2[:port2],...][/database][?options]
	// via: https://docs.mongodb.com/manual/reference/connection-string/
	connectionString := strings.Join(m.Config.Hosts, ",")
	if len(m.Config.UserName) > 0 && len(m.Config.Password) > 0 {
		connectionString = m.Config.UserName + ":" + url.QueryEscape(m.Config.Password) + "@" + connectionString
	}

	if len(m.Config.Database) > 0 {
		connectionString += "/" + m.Config.Database
	}

	if len(m.Config.AuthSource) > 0 {
		connectionString += "?authSource=" + m.Config.AuthSource
	}

	session, err := mgo.DialWithTimeout(connectionString, m.Config.Timeout)
	if err != nil {
		return err
	}

	m.Session = session
	return nil
}

func (m *Mdb) WithC(collection string, job func(*mgo.Collection) error) error {
	s := m.Session.New()
	err := job(s.DB(m.Config.Database).C(collection))
	s.Close()
	return err
}

func (self *Mdb) Upsert(collection string, selector interface{}, change interface{}) error {
	return self.WithC(collection, func(c *mgo.Collection) error {
		_, err := c.Upsert(selector, change)
		return err
	})
}

func (self *Mdb) Insert(collection string, data ...interface{}) error {
	return self.WithC(collection, func(c *mgo.Collection) error {
		return c.Insert(data...)
	})
}

func (self *Mdb) FindId(collection string, id interface{}, result interface{}) error {
	return self.WithC(collection, func(c *mgo.Collection) error {
		return c.Find(bson.M{"_id": id}).One(result)
	})
}

func (self *Mdb) FindOne(collection string, query interface{}, result interface{}) error {
	return self.WithC(collection, func(c *mgo.Collection) error {
		return c.Find(query).One(result)
	})
}

func (self *Mdb) RemoveId(collection string, id interface{}) error {
	return self.WithC(collection, func(c *mgo.Collection) error {
		return c.RemoveId(id)
	})
}
