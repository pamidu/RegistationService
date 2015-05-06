package main

import (
	"bytes"
	"code.google.com/p/gorest"
	"crypto/rand"
	"crypto/tls"
	"duov6.com/cebadapter"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
)

type User struct {
	Activated    string
	UserID       string
	EmailAddress string
	Password     string
	Token        string
	Name         string
}

type Registation struct {
	UserID       string
	EmailAddress string
	Password     string
	Name         string
}
type ResetEmail struct {
	ResetEmail string
}

type Password struct {
	EmailAddress string
	Password     string
}

//Service Definition
type RegistationService struct {
	gorest.RestService
	//gorest.RestService `root:"/tutorial/"`
	userRegistation gorest.EndPoint `method:"POST" path:"/UserRegistation/" postdata:"Registation"`
	userActivation  gorest.EndPoint `method:"GET" path:"/UserActivation/{token:string}" output:"string"`
	resetPassword   gorest.EndPoint `method:"POST" path:"/ResetPassword/" postdata:"ResetEmail"`
	passwordSet     gorest.EndPoint `method:"GET" path:"/PasswordSet/{token:string}" output:"string"`
	passwordSave    gorest.EndPoint `method:"POST" path:"/PasswordSave/" postdata:"Password"`
}

func main() {
	cebadapter.Attach("Registration", func(s bool) {
		cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
			fmt.Println("Store Configuration Successfully Loaded...")

			agent := cebadapter.GetAgent()

			agent.Client.OnEvent("globalConfigChanged.StoreConfig", func(from string, name string, data map[string]interface{}, resources map[string]interface{}) {
				cebadapter.GetLatestGlobalConfig("StoreConfig", func(data []interface{}) {
					fmt.Println("Store Configuration Successfully Updated...")
				})
			})
		})
		fmt.Println("Successfully registered in CEB")
	})

	gorest.RegisterService(new(RegistationService)) //Register our service
	http.Handle("/", gorest.Handle())
	argument := os.Args[1]
	fmt.Println(argument)
	http.ListenAndServe(":"+argument, nil)
}

//Save New user Details in DB
func RegistationDetailSave(jsondata []byte) string {
	fmt.Println("Parameter data", string(jsondata))
	url := "http://172.17.42.1:3000/com.duosoftware.com/newobject"
	fmt.Println("URL:>", url)
	result := "FALSE"
	fmt.Println("running**")
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsondata))

	req.Header.Set("securityToken", "securityToken")
	req.Header.Set("log", "log")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		result = "TRUE"
	}

	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))
	fmt.Println("running**")
	return result

}

//Register new user
func (serv RegistationService) UserRegistation(r Registation) {
	//check UserEmaill
	userCheck := Usersearchbykey(r.EmailAddress)
	userIdCheck := UserSearchbyToken(r.UserID)
	fmt.Println("\n usercheck:", userCheck, "\n", userIdCheck)
	//check userID
	if len(userCheck) == 0 && len(userIdCheck) == 0 {
		//do
		token := randToken()
		var jsonStr = []byte("{\"Object\":{\"userID\":\"" + r.UserID + "\",\"EmailAddress\":\"" + r.EmailAddress + "\",\"Name\":\"" + r.Name + "\",\"Password\":\"" + r.Password + "\",\"Token\":\"" + token + "\",\"Activated\":\"False\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
		fmt.Println(" data in Postmethod\n", "{\"Object\":{\"userID\":\""+r.UserID+"\",\"EmailAddress\":\""+r.EmailAddress+"\",\"Name\":\""+r.Name+"\",\"Password\":\""+r.Password+"\",\"Token\":\""+token+"\",\"Activated\":\"False\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
		fmt.Println("\nObjectstore\n")
		result := RegistationDetailSave(jsonStr)
		if result == "TRUE" {
			Email(r.EmailAddress, token, "Activation")
		}
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User NOT Registred\n registation process runned "))
	} else if len(userCheck) != 0 {
		//user already in
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("Already Registered "))

	} else if len(userIdCheck) != 0 {
		//userid taken
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User ID already Used "))

	} else {
		//try again
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("something wrong ... "))

	}

	/*if len(userCheck) == 0 {
		userIdCheck := UserSearchbyToken(r.UserID)

		fmt.Println("user ID", userIdCheck, "length", len(userIdCheck))
		if len(userIdCheck) == 0 {
			fmt.Println("User Not Registred.\nRegistation Process Running ")
			token := randToken()
			var jsonStr = []byte("{\"Object\":{\"userID\":\"" + r.UserID + "\",\"EmailAddress\":\"" + r.EmailAddress + "\",\"Name\":\"" + r.Name + "\",\"Password\":\"" + r.Password + "\",\"Token\":\"" + token + "\",\"Activated\":\"False\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
			fmt.Println(" data in Postmethod\n", "{\"Object\":{\"userID\":\""+r.UserID+"\",\"EmailAddress\":\""+r.EmailAddress+"\",\"Name\":\""+r.Name+"\",\"Password\":\""+r.Password+"\",\"Token\":\""+token+"\",\"Activated\":\"False\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
			fmt.Println("\nObjectstore\n")
			result := RegistationDetailSave(jsonStr)
			if result == "TRUE" {
				Email(r.EmailAddress, token, "Activation")
			}
			serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User NOT Registred\n registation process runned "))

		} else {
			serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User ID already Used "))
		}

	} else {
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("UserAlredy Registred"))
	}*/

}

//Activate user account using invitation mail send with token
func (serv RegistationService) UserActivation(token string) string {
	//check user from db
	userCheck := UserSearchbyToken(token)
	respond := ""
	fmt.Println("\n usercheck:", token, userCheck, "\n")
	if len(userCheck) == 0 {
		fmt.Println("User Not Registred.")
		respond = "User Not Found"
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User NOT Registred\n "))

	} else {

		var jsonStr = []byte("{\"Object\":{\"userID\":\"" + userCheck["userID"] + "\",\"EmailAddress\":\"" + userCheck["EmailAddress"] + "\",\"Name\":\"" + userCheck["Name"] + "\",\"Password\":\"" + userCheck["Password"] + "\",\"Token\":\"" + " " + "\",\"Activated\":\"TRUE\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
		fmt.Println("\nObjectstore\n")
		result := RegistationDetailSave(jsonStr)
		if result == "TRUE" {
		}
		respond = "UserActivated"
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User Activated"))
	}
	return respond

}

//password set
//request password reset->email send with new token ->redirect to new password set password set pase and vaidate token if token found return user email address
func (serv RegistationService) PasswordSet(token string) string {
	userCheck := UserSearchbyToken(token)
	fmt.Println("\n usercheck:", userCheck, "\n")
	if len(userCheck) == 0 {
		fmt.Println("User Not Found.")
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User Not Found"))

	} else {
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte(userCheck["EmailAddress"]))

	}
	return "set"

}

//New Password Save
func (serv RegistationService) PasswordSave(p Password) {
	userCheck := Usersearchbykey(p.EmailAddress)
	fmt.Println("\n usercheck:", userCheck, "\n")
	if len(userCheck) == 0 {
		fmt.Println("User Not Found.")
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User Not Found\n "))

	} else {

		var jsonStr = []byte("{\"Object\":{\"userID\":\"" + userCheck["UserID"] + "\",\"EmailAddress\":\"" + userCheck["EmailAddress"] + "\",\"Name\":\"" + userCheck["Name"] + "\",\"Password\":\"" + p.Password + "\",\"Token\":\"" + " " + "\",\"Activated\":\"TRUE\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
		fmt.Println("\nObjectstore\n")
		result := RegistationDetailSave(jsonStr)
		if result == "TRUE" {
			Email(p.EmailAddress, " ", "PasswordSetSuccess")
			fmt.Println("Done")
		}
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("password Reset success "))

	}
}

//request password reset
func (serv RegistationService) ResetPassword(s ResetEmail) {
	userCheck := Usersearchbykey(s.ResetEmail)
	if len(userCheck) == 0 {
		fmt.Println("User Not Found.")
		serv.ResponseBuilder().SetResponseCode(200).Write([]byte("User Not Found\n "))

	} else {
		token := randToken()
		Email(userCheck["EmailAddress"], token, "PasswordReset")
		var jsonStr = []byte("{\"Object\":{\"userID\":\"" + userCheck["UserID"] + "\",\"EmailAddress\":\"" + userCheck["EmailAddress"] + "\",\"Name\":\"" + userCheck["Name"] + "\",\"Password\":\"" + userCheck["Password"] + "\",\"Token\":\"" + token + "\",\"Activated\":\"TRUE\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
		result := RegistationDetailSave(jsonStr)
		if result == "TRUE" {
			fmt.Println("Done")
		}
	}
	serv.ResponseBuilder().SetResponseCode(200).Write([]byte("Password Reset Email Sent"))

}

//send Activation ,Passowrd Reset request and password change mail
//message contating not set properly
func Email(receiver, token string, emailtype string) {
	from := mail.Address{"", senderEmail}
	to := mail.Address{"", receiver}
	subj := ""
	body := ""

	if emailtype == "Activation" {
		subj = "DuoWorld Activation Requierd"
		body = `"<!DOCTYPE html>
<html>
<head>
 <title></title>
 <link rel=\"stylesheet\" href=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.min.css\">
 <style type=\"text/css\">
  .emailContent{
      position: relative;
      padding: 60px 0 60px 0;
      width: 869px;
      height: 493px;
      background: rgb(40, 70, 102) url('http://i58.tinypic.com/2cpp2bq.jpg') no-repeat center center;
      background-size: cover;
      color: #fff;
  }
 </style>
</head>
<body>
 <section class=\"emailContent\">
   <div class=\"row hero-content\">
    <div class=\"col-md-12 text-center\">
                    <img src=\"http://i57.tinypic.com/2qbx3c7.png\" alt=\"DuoWorld Logo\" width=\"50%\" height=\"30%\">
    </div>
    <div class=\"col-md-12 text-center\">
    <p><h2>Just one more step...</h2></p>
    <p><h4>Click the big button below to activate your DuoWorld account.</h4></p>
     <br/>
     <button class=\"btn\" style=\"width: 150px;height: 50px;font-size: 27px;background-color: aquamarine;\"><a href=\"http://duoworld.sossgrid.com:1000/UserActivation/"+token+" >Activate</a></button>
    </div>

  </div>
 </section>
</body>
</html>"` // two lines\n http://duoworld.sossgrid.com:1000/UserActivation/" + token

	} else if emailtype == "PasswordReset" {
		subj = "This is Password Reset Request "
		body = "Click To Reset Password.\n With two lines\n http://duoworld.sossgrid.com:1000/PasswordSet/" + token

	} else if emailtype == "PasswordSetSuccess" {
		subj = "This is Password set success"
		body = "password set success"
	}
	headers := make(map[string]string)
	headers["From"] = from.String()
	headers["To"] = to.String()
	headers["Subject"] = subj
	headers["Content-type"] = "text/html"

	// Setup message
	message := ""
	for k, v := range headers {
		message += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	message += "\r\n" + body
	// Connect to the SMTP Server
	servername := "173.194.65.108:465"
	host, _, _ := net.SplitHostPort(servername)
	auth := smtp.PlainAuth("",senderEmail, pwd, host)

	// TLS config
	//fmt.Print("7")
	tlsconfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:         host,
	}

	// Here is the key, you need to call tls.Dial instead of smtp.Dial
	// for smtp servers running on 465 that require an ssl connection
	// from the very beginning (no starttls)
	//fmt.Print("8")
	conn, err := tls.Dial("tcp", servername, tlsconfig)
	if err != nil {
		log.Panic(err)
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		log.Panic(err)
	}

	// Auth
	if err = c.Auth(auth); err != nil {
		log.Panic(err)
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		log.Panic(err)
	}

	if err = c.Rcpt(to.Address); err != nil {
		log.Panic(err)
	}

	// Data
	w, err := c.Data()
	if err != nil {
		log.Panic(err)
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		log.Panic(err)
	}

	err = w.Close()
	if err != nil {
		log.Panic(err)
	} else {
		fmt.Println("\nMail sent sucessfully....")
	}

	c.Quit()

}

//search using keyproperty via obgectstore
func Usersearchbykey(p string) (retdata map[string]string) {

	retdata = make(map[string]string)
	fmt.Println(p)
	url := "http://172.17.42.1:3000/com.duosoftware.com/newobject/" + p
	fmt.Println("URL:>", url)

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println(err.Error())
	}

	req.Header.Set("securityToken", "securityToken")
	req.Header.Set("log", "log")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if resp.Status == "500 Internal Server Error" || len(resp.Status) == 25 {

		var dd map[string]string
		dd = make(map[string]string)
		retdata = dd
	} else {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)

		var array map[string]interface{}
		array = make(map[string]interface{})
		_ = json.Unmarshal(body, &array)

		var data map[string]string
		data = make(map[string]string)

		for fieldName, value := range array {

			if fieldName != "__osHeaders" {
				fmt.Print(fieldName + " : ")
				fmt.Println(value.(string))
				data[fieldName] = value.(string)
			}
		}

		resp.Body.Close()
		retdata = data

	}

	return retdata

}

//search using any key via obgectstor
func UserSearchbyToken(t string) (retdata map[string]string) {

	retdata = make(map[string]string)
	url := "http://172.17.42.1:3000/com.duosoftware.com/newobject?keyword=" + t
	fmt.Println("URL:>", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println(err.Error())
	}

	req.Header.Set("securityToken", "securityToken")
	req.Header.Set("log", "log")

	client := &http.Client{}
	resp, err := client.Do(req)
	defer resp.Body.Close()
	fmt.Println(resp.Status)
	if resp.Status == "500 Internal Server Error" {
		var dd map[string]string
		dd = make(map[string]string)
		retdata = dd
	} else {
		fmt.Println("response Status:", resp.Status)
		fmt.Println("response Headers:", resp.Header)
		body, _ := ioutil.ReadAll(resp.Body)
		fmt.Println("\nresponse Body:", string(body))

		var array []map[string]interface{}
		array = make([]map[string]interface{}, 1)
		_ = json.Unmarshal(body, &array)

		var data map[string]string
		data = make(map[string]string)
		if len(array) != 0 {
			for fieldName, value := range array[0] {
				if fieldName != "__osHeaders" {
					fmt.Print(fieldName + " : ")
					fmt.Println(value.(string))
					data[fieldName] = value.(string)
				}
			}
		}

		resp.Body.Close()
		retdata = data

	}
	return retdata

}

//genarate random token
func randToken() string {
	b := make([]byte, 8)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}
