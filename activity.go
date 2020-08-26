package sendmailv2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
	"strings"
	"time"
)

type Appointment struct {
	patient string
}

// ActivityLog is the default logger for the Log Activity
var activityLog = logger.GetLogger("activity-flogo-sendmail")

// MyActivity is a stub for your Activity implementation
type sendmail struct {
	metadata *activity.Metadata
}

// NewActivity creates a new activity
func NewActivity(metadata *activity.Metadata) activity.Activity {
	return &sendmail{metadata: metadata}
}

// Metadata implements activity.Activity.Metadata
func (a *sendmail) Metadata() *activity.Metadata {
	return a.metadata
}

// Eval implements activity.Activity.Eval
func (a *sendmail) Eval(ctx activity.Context) (done bool, err error) {

	emailtype := ctx.GetInput("type");

	if emailtype == "appointment"{
		createAppointment(ctx);
	}else{
		createPrescription(ctx);
	}


	return true, nil
}

func createPrescription(ctx activity.Context){
	server := ctx.GetInput("1_smtp_server").(string)
	port := ctx.GetInput("1_smtp_port").(string)
	emailauth := ctx.GetInput("1_smtp_auth_email").(string)
	from_name := ctx.GetInput("1_smtp_sender_name").(string)
	ssl := ctx.GetInput("1_smtp_ssl").(string)
	bcc := ctx.GetInput("1_smtp_bcc_email").(string)
	apppass := ""
	email_from := emailauth
	template := ctx.GetInput("5_template_name").(string)

	if ssl != "true" {
		apppass = ctx.GetInput("1_smtp_auth_password").(string)
		email_from = ctx.GetInput("1_smtp_from_email").(string)

	}

	teste := ctx.GetInput("drugs").([][]interface{})
	//for index, element := range teste {
	//	// index is the index where we are
	//	// element is the element from someSlice for where we are
	//}

	//field := make(map[string]interface{})


	//sampleText := "";
var teste1 []Teste
	stringteste := ""

	for i := 0; i < len(teste); i++ {
		for j := 0; j < len(teste[i]); j++{
			//x := fmt.Sprintf("%v", teste[i][j])
			x:= teste[i][j]
			x = *x.(*string)
			y := x.(string)
			y = strings.ReplaceAll(y, "{", "")

			f := NewRequest([]string{""}, "medicação", "")

			errorf := f.ParseTemplate("template-teste.html", y)
			fmt.Println(errorf)
			if errorf := f.ParseTemplate(template+".html", y); errorf == nil {
				stringteste += f.body;
				fmt.Println(f.body)
			}
			//html:= "<div><strong>" + y + "</strong></div>"
			//sampleText += fmt.Sprintf(html)
			//fmt.Println(html)
			t := NewTeste(y)
			teste1 = append(teste1, t)

			fmt.Println(teste1)
		}
	}


	for i:=0; i < len(teste1) ; i++{


	}

	fmt.Println("string");
	fmt.Println(stringteste)

	log.Println(teste1);

	//for _,item:=range teste.([]interface{}) {
	//	fmt.Printf("%v", item.([]interface{})[0])
	//}

	dispensation_pin := ctx.GetInput("prescription_option_pin").(string)
	option_pin := ctx.GetInput("option_pin").(string)

	delimeter          := "**=cuf689407924327"

	ercpnt := ctx.GetInput("3_patient_contact").(string)
	from := from_name + " <" + email_from + ">";

	sampleMsg := fmt.Sprintf("From: %s\r\n", from)
	sampleMsg += fmt.Sprintf("To: %s\r\n", ercpnt)
	sampleMsg += "Subject: " + "teste" + "\r\n"
	sampleMsg += "MIME-Version: 1.0\r\n"
	sampleMsg += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter)
	sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	sampleMsg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	sampleMsg += "Content-Transfer-Encoding: 7bit\r\n"

	to := []string{""}

	if bcc == ""{
		to = []string{ercpnt}
	}else{
		to = []string{ercpnt, bcc}
	}

	templateData := struct{
		Number	string
		DismissalCode string
		RightCode string
		Text []Teste
	}{
		Number: "123456",
		DismissalCode: dispensation_pin,
		RightCode: option_pin,
		Text: teste1,
	}

	r := NewRequest([]string{ercpnt}, "medicação", "")



	error1 := r.ParseTemplate(template+".html", templateData)
	fmt.Println(error1)
	if error1 := r.ParseTemplate(template+".html", templateData); error1 == nil {
		sampleMsg += r.body

		if ssl != "true" {
			auth := smtp.PlainAuth("", emailauth, apppass, server)
			err := smtp.SendMail(server+":"+port, auth, email_from, to, []byte(sampleMsg))
			if(err != nil){
				fmt.Println(err)
				//handleError(endpoint, appointment_int_id)
			}else{
				//saveTemplateEmail(sampleMsg, endpoint_email_template, appointment_int_id)
			}
		}else{
			err := smtp.SendMail(server+":"+port, nil, email_from, to, []byte(sampleMsg))
			if(err != nil){
				fmt.Println(err)
				//handleError(endpoint, appointment_int_id)
			}else{
				//saveTemplateEmail(sampleMsg, endpoint_email_template, appointment_int_id)
			}
		}


		log.Print("done.")

	}





	//if ssl != "true" {
	//	auth := smtp.PlainAuth("", emailauth, apppass, server)
	//	err := smtp.SendMail(server+":"+port, auth, email_from, to, []byte(sampleMsg))
	//	if(err != nil){
	//		fmt.Println(err)
	//		//handleError(endpoint, appointment_int_id)
	//	}else{
	//		//saveTemplateEmail(sampleMsg, endpoint_email_template, appointment_int_id)
	//	}
	//}else{
	//	err := smtp.SendMail(server+":"+port, nil, email_from, to, []byte(sampleMsg))
	//	if(err != nil){
	//		fmt.Println(err)
	//		//handleError(endpoint, appointment_int_id)
	//	}else{
	//		//saveTemplateEmail(sampleMsg, endpoint_email_template, appointment_int_id)
	//	}
	//}

}
type Teste struct {
	teste string
}

func NewTeste(teste string) Teste{
	return Teste{
		teste: teste,
	}
}

type Prescription struct{
	number	string
	dismissalCode string
	rightCode string
}

func NewPrescription(number string, dismissalCode string, rightCode string) *Prescription{
	return &Prescription{
		number: number,
		dismissalCode: dismissalCode,
		rightCode: rightCode,
	}
}

func createAppointment(ctx activity.Context){
	//get input vars
	server := ctx.GetInput("1_smtp_server").(string)
	port := ctx.GetInput("1_smtp_port").(string)
	emailauth := ctx.GetInput("1_smtp_auth_email").(string)
	from_name := ctx.GetInput("1_smtp_sender_name").(string)
	ssl := ctx.GetInput("1_smtp_ssl").(string)
	bcc := ctx.GetInput("1_smtp_bcc_email").(string)
	apppass := ""
	email_from := emailauth

	if ssl != "true" {
		apppass = ctx.GetInput("1_smtp_auth_password").(string)
		email_from = ctx.GetInput("1_smtp_from_email").(string)

	}


	appointment := ctx.GetInput("2_appointment_name").(string)
	date := ctx.GetInput("2_appointment_date").(string)


	clinic := ctx.GetInput("2_appointment_hospital").(string)
	meet := ctx.GetInput("2_appointment_meet").(string)
	subject := ctx.GetInput("2_appointment_subject").(string)
	status := ctx.GetInput("2_appointment_status").(string)
	appointment_id := ctx.GetInput("2_appointment_id").(string)
	enddate := ctx.GetInput("2_appointment_end_date").(string)
	appointment_int_id := ctx.GetInput("2_appointment_int_id").(string)

	ercpnt := ctx.GetInput("3_patient_contact").(string)
	patient := ctx.GetInput("3_patient_name").(string)

	practitioner := ctx.GetInput("4_practitioner_name").(string)

	template := ctx.GetInput("5_template_name").(string)
	image_footer := ctx.GetInput("5_template_image_footer").(string)
	link_footer := ctx.GetInput("5_template_link_footer").(string)
	image_footer_alt := ctx.GetInput("5_template_image_footer_alt").(string)

	organizer := ctx.GetInput("6_ics_organizer").(string)
	prodid := ctx.GetInput("6_ics_prodid").(string)

	endpoint := ctx.GetInput("1_smtp_error_endpoint").(string)
	endpoint_email_template := ctx.GetInput("7_endpoint_email_template").(string)


	method := "CANCEL"
	fstatus := "CANCELLED"
	transp := "TRANSPARENT"
	if status != "cancelled" {
		method = "PUBLISH"
		fstatus = "CONFIRMED"
		transp = "OPAQUE"
	}


	date1 := time.Now()
	fdate1 := date1.Format("20060102T150405Z")

	loc, err := time.LoadLocation("Europe/Lisbon")
	layout := "2006-01-02T15:04:05.000-0700"
	fmt.Println(err);
	startDate, errd := time.Parse(layout, date)

	fenddade, errd := time.Parse(layout, enddate)



	fmt.Println(errd)

	content := "BEGIN:VCALENDAR\r"+
		"METHOD:" + method + "\r"+
		"PRODID:" + prodid + "\r"+
		"VERSION:2.0\r"+
		"X-WR-TIMEZONE:Europe/Lisbon\r" +
		"BEGIN:VTIMEZONE\r"+
		"TZID:Europe/Lisbon\r"+
		"X-LIC-LOCATION:Europe/Lisbon\r" +
		"LAST-MODIFIED:20050809T050000Z\r"+
		"BEGIN:STANDARD\r"+
		"DTSTART:20071104T020000\r"+
		"TZOFFSETFROM:+0100\r"+
		"TZOFFSETTO:+0000\r"+
		"TZNAME:WET\r"+
		"END:STANDARD\r"+
		"BEGIN:DAYLIGHT\r"+
		"DTSTART:20070311T020000\r"+
		"TZOFFSETFROM:+0000\r"+
		"TZOFFSETTO:+0100\r"+
		"TZNAME:WEST\r"+
		"END:DAYLIGHT\r"+
		"END:VTIMEZONE\r"+
		"BEGIN:VEVENT\r" +
		"DTSTAMP:" + fdate1 + "\r" +
		"UID:" + appointment_id + "\r" +
		"SEQUENCE:0\r" +
		"ORGANIZER;" + organizer + "\r" +
		"DTSTART:" + startDate.Format("20060102T150405Z") + "\r" +
		"DTEND:" + fenddade.Format("20060102T150405Z") + "\r" +
		//"DTSTART;TZID=\"Europe/Lisbon\":" + startDate.Format("20060102T150405Z") + "\r" +
		//"DTEND;TZID=\"Europe/Lisbon\":" + fenddade.Format("20060102T150405Z") + "\r" +
		"STATUS:" + fstatus + "\r" +
		"CATEGORIES:" + appointment + " " + clinic + "\r" +
		"SUMMARY:" + appointment + " " + clinic + "\r" +
		"CLASS:PUBLIC\r" +
		"TRANSP:" + transp + "\r" +
		"END:VEVENT\r" +
		"END:VCALENDAR\r"


	filename1 := CreateTempFile(content)

	//create email

	var (
		serverAddr         = server
		portNumber         = port
		tos                = ercpnt
		attachmentFilePath = filename1
		filename           = "invite.ics"
		delimeter          = "**=cuf689407924327"
	)

	from := from_name + " <" + email_from + ">";

	sampleMsg := fmt.Sprintf("From: %s\r\n", from)
	sampleMsg += fmt.Sprintf("To: %s\r\n", tos)
	sampleMsg += "Subject: " + subject + "\r\n"
	sampleMsg += "MIME-Version: 1.0\r\n"
	sampleMsg += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter)
	sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	sampleMsg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	sampleMsg += "Content-Transfer-Encoding: 7bit\r\n"

	startDate = startDate.In(loc)
	fenddade = fenddade.In(loc)


	templateData := struct {
		Name         string
		Appointment  string
		Practitioner string
		Date         string
		Hour         string
		Meet         string
		Hospital     string
		Footer       string
		Image        string
		Alt          string
	}{
		Name:         patient,
		Appointment:  appointment,
		Practitioner: practitioner,
		Date:         strconv.Itoa(startDate.Day()) + "/" + strconv.Itoa(int(startDate.Month())),
		Hour:         handlehour(startDate.Hour()) + ":" + handlehour(startDate.Minute()),
		Meet:         meet,
		Hospital:     clinic,
		Footer:       link_footer,
		Image:        image_footer,
		Alt:          image_footer_alt,
	}

	r := NewRequest([]string{ercpnt}, subject, "")
	error1 := r.ParseTemplate(template+".html", templateData)
	if error1 := r.ParseTemplate(template+".html", templateData); error1 == nil {
		sampleMsg += r.body

		sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
		sampleMsg += "Content-Type: text/calendar; charset=\"utf-8\"\r\n"
		sampleMsg += "Content-Transfer-Encoding: base64\r\n"
		sampleMsg += "Content-Disposition: attachment;filename=\"" + filename + "\"\r\n"

		rawFile, fileErr := ioutil.ReadFile(attachmentFilePath)
		if fileErr != nil {
			log.Panic(fileErr)
		}
		sampleMsg += "\r\n" + base64.StdEncoding.EncodeToString(rawFile)


		log.Println("Write content into client writter I/O")

		to := []string{tos, bcc}


		if ssl != "true" {
			auth := smtp.PlainAuth("", emailauth, apppass, serverAddr)
			err := smtp.SendMail(serverAddr+":"+portNumber, auth, email_from, to, []byte(sampleMsg))
			if(err != nil){
				fmt.Println(err)
				handleError(endpoint, appointment_int_id)
			}else{
				saveTemplateEmail(sampleMsg, endpoint_email_template, appointment_int_id)
			}
		}else{
			err := smtp.SendMail(serverAddr+":"+portNumber, nil, email_from, to, []byte(sampleMsg))
			if(err != nil){
				fmt.Println(err)
				handleError(endpoint, appointment_int_id)
			}else{
				saveTemplateEmail(sampleMsg, endpoint_email_template, appointment_int_id)
			}
		}


		log.Print("done.")

		defer os.Remove(filename)

	}
	fmt.Println(error1)
}

func CreateTempFile(serializer string) string {

	tmpFile, err := ioutil.TempFile(os.TempDir(), "*.ics")
	if err != nil {
		log.Fatal("Cannot create temporary file", err)
	}

	// Remember to clean up the file afterwards
	//defer os.Remove(tmpFile.Name())

	fmt.Println("Created File: " + tmpFile.Name())

	// Example writing to the file
	text := []byte(serializer)
	if _, err = tmpFile.Write(text); err != nil {
		log.Fatal("Failed to write to temporary file", err)
	}

	// Close the file
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	return tmpFile.Name()
}

type Request struct {
	from    string
	to      []string
	subject string
	body    string
}

func NewRequest(to []string, subject, body string) *Request {
	return &Request{
		to:      to,
		subject: subject,
		body:    body,
	}
}

func (r *Request) ParseTemplate(templateFileName string, data interface{}) error {
	t, err := template.ParseFiles(templateFileName)
	if err != nil {
		return err
	}
	buf := new(bytes.Buffer)
	if err = t.Execute(buf, data); err != nil {
		return err
	}
	r.body = buf.String()
	return nil
}

func handleError(endpoint string, id string) {
	fmt.Println("Init retry update")

	requestBody, err1 := json.Marshal(map[string]string{
	})
	if err1 == nil{
		fmt.Println(err1)
	}
	response, err := http.Post(endpoint + "/" + id, "application/json", bytes.NewBuffer(requestBody))
	if err == nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
	fmt.Println("Terminating retry update")
}


func saveTemplateEmail(text string, endpoint string, id string){
	requestBody, err1 := json.Marshal(map[string]string{
		"text" : text,
	})
	if err1 != nil{
		log.Fatalln(err1)
	}
	response, err := http.Post(endpoint + "/" + id, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}


}

func handlehour(number int) (formatted string){
	formatted = strconv.Itoa(number)
	if number == 0{
		formatted = "00"
	}else{
		text := strconv.Itoa(number)
		if len(text)  == 1{
			formatted = "0" + text
		}
	}
	return formatted
}

