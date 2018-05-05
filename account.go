package cronsun

import (
	"time"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	Coll_Account = "account"
)

type Account struct {
	ID       bson.ObjectId `bson:"_id" json:"id"`
	Role     Role          `bson:"role" json:"role"`
	Email    string        `bson:"email" json:"email"`
	Password string        `bson:"password" json:"password"`
	Salt     string        `bson:"salt" json:"salt"`
	Status   UserStatus    `bson:"status" json:"status"`
	Session  string        `bson:"session" json:"-"`
	// If true, role and status are unchangeable, email and password can be change by it self only.
	Unchangeable bool      `bson:"unchangeable" json:"-"`
	CreateTime   time.Time `bson:"createTime" json:"createTime"`
}

type Role int

const (
	Administrator Role = 1
	Developer     Role = 2
	Reporter      Role = 3
)

func (r Role) Defined() bool {
	switch r {
	case Administrator, Developer, Reporter:
		return true
	}
	return false
}

func (r Role) String() string {
	switch r {
	case Administrator:
		return "Administrator"
	case Developer:
		return "Developer"
	case Reporter:
		return "Reporter"
	}
	return "Undefined"
}

type UserStatus int

const (
	UserBanned  UserStatus = -1
	UserActived UserStatus = 1
)

func (s UserStatus) Defined() bool {
	switch s {
	case UserBanned, UserActived:
		return true
	}
	return false
}

func GetAccounts(query bson.M) (list []Account, err error) {
	err = mgoDB.WithC(Coll_Account, func(c *mgo.Collection) error {
		return c.Find(query).All(&list)
	})
	return
}

func GetAccountByEmail(email string) (u *Account, err error) {
	err = mgoDB.FindOne(Coll_Account, bson.M{"email": email}, &u)
	return
}

func CreateAccount(u *Account) error {
	u.ID = bson.NewObjectId()
	u.CreateTime = time.Now()
	return mgoDB.Insert(Coll_Account, u)

}

func UpdateAccount(query bson.M, change bson.M) error {
	return mgoDB.WithC(Coll_Account, func(c *mgo.Collection) error {
		return c.Update(query, bson.M{"$set": change})
	})
}

func BanAccount(email string) error {
	return mgoDB.WithC(Coll_Account, func(c *mgo.Collection) error {
		return c.Update(bson.M{"email": email}, bson.M{"$set": bson.M{"status": UserBanned}})
	})
}

func EnsureAccountIndex() error {
	return mgoDB.WithC(Coll_Account, func(c *mgo.Collection) error {
		return c.EnsureIndex(mgo.Index{
			Key:    []string{"email"},
			Unique: true,
		})
	})
}
