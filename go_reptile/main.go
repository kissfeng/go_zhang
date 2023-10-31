package main

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os"
)

// MySQL 数据库配置
const (
	dbHost     = "locahost"
	dbUser     = "root"
	dbPassword = "123456!"
	dbName     = "test_db"
)

// 分块查询配置
const chunkSize = 200 // 每个分块的大小
// 导入csv
func main() {
	// 连接到 MySQL 数据库
	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbHost, dbName))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// 查询商品表的总行数
	var totalRows int
	countQuery := "SELECT COUNT(a.id) FROM mpn_test a LEFT JOIN products b ON a.mpn=b.mpn WHERE b.mpn IS NULL"
	err = db.QueryRow(countQuery).Scan(&totalRows)
	if err != nil {
		log.Fatal(err)
	}

	// 计算分块数
	numChunks := (totalRows + chunkSize - 1) / chunkSize

	// 创建 CSV 文件
	file, err := os.Create("output.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	headers := []string{"mpn", "ProductList"}
	err = writer.Write(headers)
	if err != nil {
		log.Fatal(err)
	}

	// 遍历每个分块查询
	for chunkIndex := 0; chunkIndex < numChunks; chunkIndex++ {
		// 计算当前分块的起始行和结束行
		startRow := chunkIndex * chunkSize

		// 分块查询商品表的图片地址
		query := fmt.Sprintf("SELECT a.mpn FROM mpn_test a LEFT JOIN products b ON a.mpn=b.mpn WHERE b.mpn IS NULL ORDER BY a.id LIMIT %d, %d", startRow, chunkSize)
		rows, err := db.Query(query)
		if err != nil {
			log.Fatal(err)
		}
		defer func(rows *sql.Rows) {
			err := rows.Close()
			if err != nil {

			}
		}(rows)

		var urls []string
		for rows.Next() {
			var mpn string
			if err := rows.Scan(&mpn); err != nil {
				panic(err)
			}
			urls = append(urls, mpn)
		}
		if err := rows.Err(); err != nil {
			panic(err)
		}
		newProducts := QueryDetails(urls)
		//fmt.Println(newProducts)
		// 遍历图片地址并上传到 OSS，同时更新数据库表
		var datas [][]interface{}
		for _, row := range newProducts {
			if row != nil {
				mpn := row[0].(string)
				// 其他操作
				productDict := row[1]
				data := []interface{}{mpn, productDict}
				datas = append(datas, data)
			} else {
				continue
			}
		}

		for _, data := range datas {
			mpn := data[0].(string)
			productBytes := data[1].([]uint8)
			productDict := string(productBytes)
			newData := []string{mpn, productDict}
			err := writer.Write(newData)
			if err != nil {
				log.Fatal(err)
			}
		}

		writer.Flush()

		if err := writer.Error(); err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Chunk %d/%d processed\n", chunkIndex+1, numChunks)

	}

	fmt.Println("CSV file created successfully!")
}

//导入数据库
//func main() {
//	// 连接到 MySQL 数据库
//	db, err := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s", dbUser, dbPassword, dbHost, dbName))
//	if err != nil {
//		panic(err)
//	}
//	defer func(db *sql.DB) {
//		err := db.Close()
//		if err != nil {
//		}
//	}(db)
//	// 查询商品表的总行数
//	var totalRows int
//	countQuery := "SELECT COUNT(a.id) FROM mpn_test a LEFT JOIN products b ON a.mpn=b.mpn WHERE b.mpn IS NULL"
//	err = db.QueryRow(countQuery).Scan(&totalRows)
//	if err != nil {
//		panic(err)
//	}
//	// 计算分块数
//	numChunks := (totalRows + chunkSize - 1) / chunkSize
//
//	// 遍历每个分块查询
//	for chunkIndex := 0; chunkIndex < numChunks; chunkIndex++ {
//		// 计算当前分块的起始行和结束行
//		startRow := chunkIndex * chunkSize
//		// endRow := (chunkIndex + 1) * chunkSize
//		// 分块查询商品表的图片地址
//		query := fmt.Sprintf("SELECT a.mpn FROM mpn_test a LEFT JOIN products b ON a.mpn=b.mpn WHERE b.mpn IS NULL ORDER BY a.id LIMIT %d, %d", startRow, chunkSize)
//		rows, err := db.Query(query)
//		if err != nil {
//			panic(err)
//		}
//		defer func(rows *sql.Rows) {
//			err := rows.Close()
//			if err != nil {
//
//			}
//		}(rows)
//		var urls []string
//		for rows.Next() {
//			var mpn string
//			if err := rows.Scan(&mpn); err != nil {
//				panic(err)
//			}
//			urls = append(urls, mpn)
//		}
//		if err := rows.Err(); err != nil {
//			panic(err)
//		}
//		newProducts := QueryDetails(urls)
//		//fmt.Println(newProducts)
//		// 遍历图片地址并上传到 OSS，同时更新数据库表
//		var datas [][]interface{}
//		for _, row := range newProducts {
//			if row != nil {
//				mpn := row[0].(string)
//				// 其他操作
//				productDict := row[1]
//				data := []interface{}{mpn, productDict}
//				datas = append(datas, data)
//			} else {
//				continue
//			}
//		}
//		// 批量插入数据
//		stmt, err := db.Prepare("INSERT INTO icnet_products (mpn, productList) VALUES (?,?) ON DUPLICATE KEY UPDATE mpn = VALUES(mpn)")
//		if err != nil {
//			panic(err)
//		}
//		defer func(stmt *sql.Stmt) {
//			err := stmt.Close()
//			if err != nil {
//
//			}
//		}(stmt)
//		for _, data := range datas {
//			_, err := stmt.Exec(data[0], data[1])
//			if err != nil {
//				panic(err)
//			}
//		}
//
//		fmt.Printf("Chunk %d/%d processed\n", chunkIndex+1, numChunks)
//	}
//
//	fmt.Println("Images uploaded to OSS and updated in the database successfully!")
//
//}

//func main() {
//	url := []string{
//		"3352T-1-103LF",
//	}
//	results := QueryDetails(url)
//	fmt.Println(results)
//}
