package utils

import (
	"fmt"
	"math/big"
	"math/rand"
	"net"
	"strconv"
	"time"

	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

func CurTimeStr() string {
	timeTemplate1 := "2006-01-02 15:04"
	res1 := time.Now().Format(timeTemplate1)
	return res1
}

func CurTimeInt() int {
	timestr := fmt.Sprintf("%v", time.Now().Unix())
	timeint, err := strconv.Atoi(timestr)
	if err != nil {
		Logger.Errorln(err)
	}
	return timeint
}

func Iplist() []string {
	var ips []string
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		Logger.Errorln(err)
		return ips
	}

	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}
	return ips
}

func Random(min, max int) int {
	if min == max {
		return min
	}
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(max-min) + min
}

func InetNtoA(ips int) string {
	ip := int64(ips)
	return fmt.Sprintf("%d.%d.%d.%d",
		byte(ip>>24), byte(ip>>16), byte(ip>>8), byte(ip))
}

func InetAtoN(ip string) int {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return int(ret.Int64())
}

func InetAtoN64(ip string) int64 {
	ret := big.NewInt(0)
	ret.SetBytes(net.ParseIP(ip).To4())
	return ret.Int64()
}

func SqlCreate() {
	db, _ := sql.Open("sqlite3", "./usdn.db")
	defer db.Close()
	//checkErr(err)
	sql_table := `create table macinfo ( list text);`
	db.Exec(sql_table)
}

// func SqlSelect() (string, error) {
// 	var k string
// 	db, err := sql.Open("sqlite3", "./usdn.db")
// 	if err != nil {
// 		return k, err
// 	}
// 	defer db.Close()
// 	stmt, err := db.Prepare("select list from macinfo;")
// 	if err != nil {
// 		return k, err
// 	}
// 	err = stmt.QueryRow().Scan(&k)
// 	if err != nil {
// 		return k, err
// 	}
// 	return k, nil
// }

func SqlSelect() (string, error) {
	var list string
	db, err := sql.Open("sqlite3", "./usdn.db")
	if err != nil {
		return list, err
	}
	defer db.Close()
	rows, err := db.Query("select list from macinfo;")
	if err != nil {
		return list, err
	}
	defer rows.Close()

	var cnt int
	for rows.Next() {
		cnt += 1
		err = rows.Scan(&list)
		if err != nil {
			return list, err
		}
	}
	if cnt == 0 {
		if err := SqlInsert(""); err != nil {
			return list, err
		}
	}
	return list, nil
}

func SqlInsert(data string) error {
	db, err := sql.Open("sqlite3", "./usdn.db")
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, err := db.Prepare("insert into macinfo(list) values(?)")
	if err != nil {
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(data)
	if err != nil {
		return err
	}

	return nil
}

func SqlUpdate(k string) error {
	db, err := sql.Open("sqlite3", "./usdn.db")
	if err != nil {
		return err
	}
	defer db.Close()
	stmt, _ := db.Prepare("update macinfo set list=?")
	_, err = stmt.Exec(k)
	if err != nil {
		return err
	}
	return nil
}
