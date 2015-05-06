package main

import (
	"bytes"
	"code.google.com/p/gorest"
	"crypto/rand"
	"crypto/tls"
	//"duov6.com/applib"
	"duov6.com/cebadapter"
	"duov6.com/common"
	"encoding/json"
	"fmt"
	"io/ioutil"
	//"log"
	//"duov6.com/config"
	"duov6.com/objectstore/client"
	"duov6.com/term"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"os"
)

type User struct {
	UserID          string
	EmailAddress    string
	Name            string
	Password        string
	ConfirmPassword string
	Active          bool
}

type Registation struct {
	UserID          string
	EmailAddress    string
	Password        string
	Name            string
	ConfirmPassword string
}
type ResetEmail struct {
	ResetEmail string
}
type AuthHandler struct {
	//Config AuthConfig
}
type Password struct {
	EmailAddress string
	Password     string
}
type Login struct {
	EmailAddress string
	Password     string
}
type ActivationEmail struct {
	EmailAddress string
	Token        string
}

type AuthConfig struct {
	Cirtifcate    string
	PrivateKey    string
	Https_Enabled bool
	StoreID       string
	Smtpserver    string
	Smtpusername  string
	Smtppassword  string
	UserName      string
	Password      string
}

func newAuthHandler() *AuthHandler {
	authhld := new(AuthHandler)
	//authhld.Config = GetConfig()
	return authhld
}

var Config AuthConfig

//Service Definition
type RegistationService struct {
	gorest.RestService
	//gorest.RestService `root:"/tutorial/"`
	userRegistation gorest.EndPoint `method:"POST" path:"/UserRegistation/" postdata:"Registation"`
	userActivation  gorest.EndPoint `method:"GET" path:"/UserActivation/{token:string}" output:"string"`
	login           gorest.EndPoint `method:"POST" path:"/Login/" postdata:"Login"`
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

//Register new user
func (serv RegistationService) UserRegistation(r Registation) {
	var user User
	user.Active = false
	user.ConfirmPassword = r.ConfirmPassword
	user.EmailAddress = r.EmailAddress
	user.Name = r.Name
	user.Password = r.Password
	//user.UserID = r.UserID
	fmt.Println("SAVE USER\n\n\n")
	//Save user Method
	res := SaveUser(user)
	fmt.Println(res)
	serv.ResponseBuilder().SetResponseCode(200).Write([]byte("done..."))

	//check UserEmaill
	/*userCheck := Usersearchbykey(r.EmailAddress)
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

	}*/

}

//Save user using Authlib

func SaveUser(u User) string {
	term.Write("SaveUser saving user  "+u.Name, term.Debug)
	respond := ""
	token := randToken()
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().BySearching(u.EmailAddress).Ok()

	if err == "" {
		var uList []User
		err := json.Unmarshal(bytes, &uList)
		if err == nil || bytes == nil {
			//new user
			if len(uList) == 0 {
				u.UserID = common.GetGUID()
				term.Write("SaveUser saving user"+u.Name+" New User "+u.UserID, term.Debug)
				client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
				respond = "true"
				//save Activation mail details
				//EmailAddress and Token
				//EmailAddress KeyProperty
				var Activ ActivationEmail
				Activ.EmailAddress = u.EmailAddress
				Activ.Token = token
				client.Go("ignore", "com.duosoftware.com", "Activation").StoreObject().WithKeyField("EmailAddress").AndStoreOne(Activ).Ok()
				Email(token, u.EmailAddress, "Activation")

			} else {
				//Alredy in  Registerd user
				term.Write("User Already Registerd  #"+err.Error(), term.Error)

				/*u.UserID = uList[0].UserID
				u.Password = uList[0].Password
				u.ConfirmPassword = uList[0].Password
				term.Write("SaveUser saving user  "+u.Name+" Update User "+u.UserID, term.Debug)
				client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
				respond = "true"
				var Activ ActivationEmail
				Activ.EmailAddress = u.EmailAddress
				Activ.Token = token
				client.Go("ignore", "com.duosoftware.com", Activ).StoreObject().WithKeyField("EmailAddress").AndStoreOne(Activ).Ok()

				TokenEmailSave(token, u.EmailAddress)*/
			}
		} else {
			term.Write("SaveUser saving user store Error #"+err.Error(), term.Error)
			respond = "false"

		}
	} else {
		term.Write("SaveUser saving user fetech Error #"+err, term.Error)
		respond = "false"

	}
	u.Password = "*****"
	u.ConfirmPassword = "*****"
	return respond
}

//Activate user account using invitation mail send with token
func (serv RegistationService) UserActivation(token string) string {
	respond := ""
	//check user from db
	bytes, err := client.Go("ignore", "com.duosoftware.com", "Activation").GetOne().BySearching(token).Ok()
	if err == "" {
		var uList []User
		err := json.Unmarshal(bytes, &uList)
		if err == nil || bytes == nil {
			//new user
			if len(uList) == 0 {

				term.Write("User Not Found", term.Debug)

			} else {
				var u User
				u.UserID = uList[0].UserID
				u.Password = uList[0].Password
				u.Active = true
				u.ConfirmPassword = uList[0].Password
				u.Name = uList[0].Name
				u.EmailAddress = uList[0].EmailAddress

				term.Write("Activate User  "+u.Name+" Update User "+u.UserID, term.Debug)
				client.Go("ignore", "com.duosoftware.auth", "users").StoreObject().WithKeyField("EmailAddress").AndStoreOne(u).Ok()
				respond = "true"
				var Activ ActivationEmail
				Activ.EmailAddress = u.EmailAddress
				Activ.Token = ""
				client.Go("ignore", "com.duosoftware.com", "Activation").StoreObject().WithKeyField("EmailAddress").AndStoreOne(Activ).Ok()

				Email(u.EmailAddress, Activ.Token, "Activated")
				respond = "Success"
			}
		}

	} else {
		term.Write("Activation Fail ", term.Debug)

	}

	return respond

}

//send Activation ,Passowrd Reset request and password change mail
//message contating not set properly
func Email(receiver, token string, emailtype string) string {
	res := "FALSE"
	from := mail.Address{"", "pamidu@duosoftware.com"}
	to := mail.Address{"", receiver}
	subj := ""
	body := ""

	if emailtype == "Activation" {
		subj = "DuoWorld Activation Requierd"
		body = "<html><head> <title></title> <link rel=\"stylesheet\" href=\"https://maxcdn.bootstrapcdn.com/bootstrap/3.3.4/css/bootstrap.min.css\"> </head><body><section style=\"position: relative;padding: 60px 0 60px 0;width: 869px;height: 493px;background: rgb(40, 70, 102) url('http://i58.tinypic.com/2cpp2bq.jpg') no-repeat center center; background-size: cover;color: #fff;\"><div class=\"row hero-conten\"> <div class=\"col-md-12 text-center\"><img src=\"http://i57.tinypic.com/2qbx3c7.png\" alt=\"DuoWorld Logo\" width=\"50%\" height=\"30%\"></div> <div class=\"col-md-12 text-center\"> <p><h2>Just one more step...</h2></p> <p><h4>Click the Activate button below to activate your DuoWorld account.</h4></p><br/>  <button class=\"btn\" style=\"width: 100px;height: 30px;font-size: 27px;background-color: aquamarine;\"> <a href=\"http://duoworld.sossgrid.com:1000/UserActivation/" + token + "\">Activate</a></button></div></div></section></body></html>"

	} else if emailtype == "Activated" {
		subj = "This is Activated "
		body = "Click To Reset Password.\n With two line"

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
	auth := smtp.PlainAuth("", "pamidu@duosoftware.com", "DuoS@123", host)

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
		fmt.Println(err.Error())
	}

	c, err := smtp.NewClient(conn, host)
	if err != nil {
		fmt.Println(err.Error())
	}
	// Auth
	if err = c.Auth(auth); err != nil {
		fmt.Println(err.Error())
	}

	// To && From
	if err = c.Mail(from.Address); err != nil {
		fmt.Println(err.Error())
	}

	if err = c.Rcpt(to.Address); err != nil {
		fmt.Println(err.Error())
	}

	// Data
	w, err := c.Data()
	if err != nil {
		fmt.Println(err.Error())
	}

	_, err = w.Write([]byte(message))
	if err != nil {
		fmt.Println(err.Error())
	}

	err = w.Close()
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("\nMail sent sucessfully....")
		res = "TRUE"
	}

	c.Quit()
	return res

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
func TokenEmailSave(token, email string) {
	var jsonStr = []byte("{\"Object\":{\"EmailAddress\":\"" + email + "\",\"Token\":\"" + token + "\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
	fmt.Println(" data in Postmethod\n", "{\"Object\":{\"EmailAddress\":\""+email+"\",\"Token\":\""+token+"\"}, \"Parameters\":{\"KeyProperty\":\"EmailAddress\"}}")
	fmt.Println("\nObjectstore\n")
	result := RegistationDetailSave(jsonStr)
	if result == "TRUE" {
		res := Email(email, token, "Activation")
		if res == "TRUE" {

		}
	} else {

	}

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
func (serv RegistationService) Login(l Login) {
	bytes, err := client.Go("ignore", "com.duosoftware.auth", "users").GetOne().BySearching(l.EmailAddress).Ok()

	if err == "" {
		if bytes != nil {
			var uList []User
			err := json.Unmarshal(bytes, &uList)

			if err == nil && len(uList) != 0 {
				if uList[0].Password == l.Password && uList[0].EmailAddress == l.EmailAddress {
					serv.ResponseBuilder().SetResponseCode(200).Write([]byte(uList[0].Name))
					//term.Write("password incorrect", term.Error)
				} else {
					serv.ResponseBuilder().SetResponseCode(201).Write([]byte("Password Wrong "))
				}
			} else {
				if err != nil {
					term.Write("Login  user Error "+err.Error(), term.Error)
				}
			}
		}
	} else {
		term.Write("Login  user  Error "+err, term.Error)
		serv.ResponseBuilder().SetResponseCode(201).Write([]byte(err))
	}

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
