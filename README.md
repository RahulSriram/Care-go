# Care-go
Golang server implementation for <a href="https://github.com/RahulSriram/Care">Care</a>

##Installation

###Dependencies
`go get github.com/go-sql-driver/mysql`

###Setting up
`go get github.com/RahulSriram/Care-go`

Ignore errors about `createSmsCode()` and `createDonationCode()`

After downloading, uncomment `createSmsCode()` and `createDonationCode()` in `server.go` and add your own random unique number generation algorithms in them

if you want something quick for testing, try this

```
import "time"

func createSmsCode(msgType string) string {
	return time.Now().Unix()
}

func createDonationCode(input string) string {
	return input
}
```

##Database structure
`mysql> use Care;`

`mysql> show tables;`

| Tables_in_Care |
|:--------------:|
| SmsRequest     |
| Transactions   |
| Users          |

###Transactions
`mysql> desc Transactions;`

| Field       | Type        | Null | Key | Default | Extra |
|:-----------:|:-----------:|:----:|:---:|:-------:|:-----:|
| donationId  | varchar(40) | NO   | PRI | NULL    |       |
| timestamp   | varchar(19) | NO   |     | NULL    |       |
| fromNumber  | varchar(14) | NO   |     | NULL    |       |
| toNumber    | varchar(14) | NO   |     | 0       |       |
| items       | text        | NO   |     | NULL    |       |
| status      | text        | NO   |     | NULL    |       |
| description | text        | NO   |     | NULL    |       |

###SmsRequest
`mysql> desc SmsRequest;`

| Field      | Type        | Null | Key | Default | Extra |
|:----------:|:-----------:|:----:|:---:|:-------:|:-----:|
| number     | varchar(14) | NO   |     | NULL    |       |
| code       | text        | NO   |     | NULL    |       |
| type       | varchar(6)  | NO   |     | NULL    |       |
| isCodeSent | varchar(1)  | NO   |     | n       |       |

###Users
`mysql> desc Users;`

| Field     | Type        | Null | Key | Default | Extra |
|:---------:|:-----------:|:----:|:---:|:-------:|:-----:|
| id        | text        | NO   |     | NULL    |       |
| number    | varchar(14) | NO   | PRI | NULL    |       |
| name      | text        | NO   |     | NULL    |       |
| latitude  | float(12,8) | NO   |     | NULL    |       |
| longitude | float(12,8) | NO   |     | NULL    |       |
