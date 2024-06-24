package service

import (
	"fmt"
	"net/url"

	geeOrm "github.com/gee-coder/gee/orm"
	_ "github.com/go-sql-driver/mysql"
)

// 待办 删除前缀拼接表名，使用第一个tag名代表表名
type User struct {
	Id       int64  `geeorm:"id,auto_increment"`
	UserName string `geeorm:"user_name"`
	Password string `geeorm:"password"`
	Age      int    `geeorm:"age"`
}

// dsn := "username:password@tcp(host:port)/dbname"
var dataSourceName = fmt.Sprintf("root:666666@tcp(localhost:3306)/blog?charset=utf8&loc=%s&parseTime=true", url.QueryEscape("Asia/Shanghai"))

func SaveUser() {
	db := geeOrm.Open("mysql", dataSourceName)
	defer func(db *geeOrm.GeeDb) {
		err := db.Close()
		if err != nil {
			panic(err)
		}
	}(db)
	db.Prefix = "blog_"
	user := &User{
		UserName: "geecoder",
		Password: "123456",
		Age:      30,
	}
	id, _, err := db.New(&User{}).Insert(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

}

func SaveUserBatch() {
	db := geeOrm.Open("mysql", dataSourceName)
	db.Prefix = "blog_"
	user := &User{
		UserName: "geecoder22",
		Password: "12345612",
		Age:      54,
	}
	user1 := &User{
		UserName: "geecoder11",
		Password: "123456111",
		Age:      12,
	}
	var users []any
	users = append(users, user, user1)
	id, _, err := db.New(&User{}).InsertBatch(users)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

func UpdateUser() {
	db := geeOrm.Open("mysql", dataSourceName)
	db.Prefix = "blog_"
	// id, _, err := db.New().Where("id", 1006).Where("age", 54).Update(user)
	// 单个插入
	user := &User{
		UserName: "geecoder",
		Password: "123456",
		Age:      30,
	}
	id, _, err := db.New(&User{}).Insert(user)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	// 批量插入
	var users []any
	users = append(users, user)
	id, _, err = db.New(&User{}).InsertBatch(users)
	if err != nil {
		panic(err)
	}
	fmt.Println(id)
	// 更新
	id, _, err = db.
		New(&User{}).
		Where("id", 1006).
		UpdateParam("age", 100).
		Update()
	// 查询单行数据
	err = db.New(&User{}).
		Where("id", 1006).
		Or().
		Where("age", 30).
		SelectOne(user, "user_name")
	// 查询多行数据
	users, err = db.New(&User{}).Select(&User{})
	if err != nil {
		panic(err)
	}
	for _, v := range users {
		u := v.(*User)
		fmt.Println(u)
	}

	if err != nil {
		panic(err)
	}
	fmt.Println(id)

	db.Close()
}

func SelectOne() {
	db := geeOrm.Open("mysql", dataSourceName)
	db.Prefix = "blog_"
	user := &User{}
	err := db.New(user).
		Where("id", 1006).
		Or().
		Where("age", 30).
		SelectOne(user, "user_name")
	if err != nil {
		panic(err)
	}
	fmt.Println(user)

	db.Close()
}

func Select() {
	db := geeOrm.Open("mysql", dataSourceName)
	db.Prefix = "blog_"
	user := &User{}
	users, err := db.New(user).Where("id", 1000).Order("id", "asc", "age", "desc").Select(user)
	if err != nil {
		panic(err)
	}
	for _, v := range users {
		u := v.(*User)
		fmt.Println(u)
	}
	db.Close()
}

func Count() {
	db := geeOrm.Open("mysql", dataSourceName)
	db.Prefix = "blog_"
	user := &User{}
	count, err := db.New(user).Count()
	if err != nil {
		panic(err)
	}
	fmt.Println(count)
	db.Close()
}
