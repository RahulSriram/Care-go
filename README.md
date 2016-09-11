# Care-go
Golang server implementation for https://github.com/RahulSriram/Care

##Database structure
`mysql> show tables`

| Tables_in_Care |
|:--------------:|
| SmsRequest     |
| Transactions   |
| Users          |

###Transactions
`mysql> desc Transactions;`

| Field         | Type        | Null | Key | Default | Extra |
|:-------------:|:-----------:|:----:|:---:|:-------:|:-----:|
| donationId    | varchar(40) | NO   | PRI | NULL    |       |
| timestamp     | varchar(19) | NO   |     | NULL    |       |
| number        | varchar(14) | NO   |     | NULL    |       |
| items         | text        | NO   |     | NULL    |       |
| status        | text        | NO   |     | NULL    |       |
| description   | text        | NO   |     | NULL    |       |

###SmsRequest
`mysql> desc SmsRequest;`

| Field      | Type        | Null | Key | Default | Extra |
|:----------:|:-----------:|:----:|:---:|:-------:|:-----:|
| number     | varchar(14) | NO   |     | NULL    |       |
| code       | varchar(6)  | NO   |     | NULL    |       |
| type       | varchar(3)  | NO   |     | NULL    |       |
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
