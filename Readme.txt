********build ******



go build RegistationService.go




******RUN FILE **************



./registationService portnumber 






*******USER REGISTATION ******



POST method Sample Data


{"UserID":"000002","EmailAddress":"pamidu@ABC.com","Name":"Pamidu","Password":"admin"}







*******PASSWORD RESET*******



POST method Data


{"ResetEmail":"pamidu@CVB.com"}








*******USER ACTIVATION******


POST method Data


{"Token":"f3c3ce96a82dd51e"}








*******PASSWORD RESET REQUEST *********


POST method data 


{"ResetEmail":"pamidu@ABC.com"}





*******PASSWORD SET METHOD *************


validate token and identify user account


POST method data 


{"Token":"a1fcf36b47793722"}







**********PASSWORD RESET ************


POST method data 


{"EmailAddress":"pamidu@ABC.com","Password":"CHANGE"}


