package main

import(
	"database/sql"
	"time"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
)

func createDatabase(user string,pass string,server string,dbname string)(*sql.DB,error){
	db,err:=sql.Open("mysql",user+":"+pass+"@/"+dbname)
	if err!=nil{
		return nil,err
	}
	db.SetConnMaxLifetime(time.Minute * 3)
	return db,nil
}

func insertIntoDB(database *sql.DB,query string) *sql.Stmt{
	statement,err:=database.Prepare(query)
	if err!=nil{
		fmt.Println(err)
		statement.Close()
		return nil
	}
	return statement
}

func insertGuild(database *sql.DB,id string,name string)bool{
	query:="INSERT INTO guilds(id,name) VALUES(%s,%s) ON DUPLICATE KEY UPDATE name=VALUES(name)"
	statement:=insertIntoDB(database,query)
	_,err:=statement.Exec(id,name)
	if err!=nil{
		fmt.Println(err)
		return false
	}
	return true
}

func insertChannel(database *sql.DB,id string,name string,id_guild string)bool{
	query:="INSERT INTO channels(id,name,id_guild) VALUES(%s,%s,%s) ON DUPLICATE KEY UPDATE name=VALUES(name)"
	statement:=insertIntoDB(database,query)
	_,err:=statement.Exec(id,name,id_guild)
	if err!=nil{
		fmt.Println(err)
		return false
	}
	return true
}

func insertUser(database *sql.DB,id string,name string)bool{
	query:="INSERT INTO users(id,name) VALUES(%s,%s) ON DUPLICATE KEY UPDATE name=VALUES(name)"
	statement:=insertIntoDB(database,query)
	_,err:=statement.Exec(id,name)
	if err!=nil{
		fmt.Println(err)
		return false
	}
	return true
}

func insertMessage(database *sql.DB,user_id string,message string,id_channel string)bool{
	query:="INSERT INTO messages(user_id,message,id_channel) VALUES(%s,%s,NOW(),%s)"
	statement:=insertIntoDB(database,query)
	_,err:=statement.Exec(user_id,message,id_channel)
	if err!=nil{
		fmt.Println(err)
		return false
	}
	return true
}

func insertAttachment(database *sql.DB,user_id string,attachment string,channel_id string)bool{
	query:="INSERT INTO attachments(user_id,attachment,time,channel_id) VALUES(%s,%s,NOW(),%s)"
	statement:=insertIntoDB(database,query)
	_,err:=statement.Exec(user_id,attachment,channel_id)
	if err!=nil{
		fmt.Println(err)
		return false
	}
	return true
}
