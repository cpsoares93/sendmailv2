package sendmailv2

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/TIBCOSoftware/flogo-lib/core/activity"
	"github.com/TIBCOSoftware/flogo-lib/logger"
	"github.com/spf13/cast"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"strconv"
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

	emailType := ctx.GetInput("1_a_type")

	if emailType == "appointment" {
		output, success := createAppointment(ctx)
		ctx.SetOutput("email", output)
		ctx.SetOutput("sent", success)
	} else {
		output, success := createPrescription(ctx)
		ctx.SetOutput("email", output)
		ctx.SetOutput("sent", success)

	}

	return true, nil
}

func createPrescription(ctx activity.Context) (email string, success bool) {
	output := ""

	server := ctx.GetInput("1_b_smtp_server").(string)
	port := ctx.GetInput("1_c_smtp_port").(string)
	emailAuth := ctx.GetInput("1_e_smtp_auth_email").(string)
	fromName := ctx.GetInput("1_g_smtp_sender_name").(string)
	ssl := ctx.GetInput("1_d_smtp_ssl").(string)
	bcc := ctx.GetInput("1_i_smtp_bcc_email").(string)
	password := ""
	emailFrom := emailAuth
	endpoint := ctx.GetInput("1_j_smtp_error_endpoint").(string)
	iterateTemplate := ctx.GetInput("5_h_prescription_template_drugs").(string)
	footerTemplate := ctx.GetInput("5_i_template_footer").(string)
	contentTemplate := ctx.GetInput("5_g_prescription_template_content").(string)

	if ssl != "true" {
		password = ctx.GetInput("1_f_smtp_auth_password").(string)
		emailFrom = ctx.GetInput("1_h_smtp_from_email").(string)
	}

	prescriptionContent := ctx.GetInput("5_f_prescription_drugs").([][]interface{})

	subject := ctx.GetInput("1_l_subject").(string)

	tableDrugs := ""

	requestId := ""
	index := 1

	data := struct {
		Index       string
		Name        string
		Dosage      string
		Pharmform   string
		Package     string
		Dosagedrug  string
		Quantity    string
		Lowest      string
		Expiration  string
		Instruction string
	}{
		Index:       strconv.Itoa(index),
		Name:        "",
		Dosage:      "",
		Pharmform:   "",
		Package:     "",
		Dosagedrug:  "",
		Quantity:    "",
		Lowest:      "",
		Expiration:  "",
		Instruction: "",
	}

	for i := 0; i < len(prescriptionContent); i++ {
		prescId := prescriptionContent[i][0]
		prescId = *prescId.(*string)

		if prescId.(string) != requestId {

			prescRequest := NewRequest([]string{""}, "Prescription", "")

			if requestId != "" {
				errorPresc := prescRequest.ParseTemplate(iterateTemplate+".html", data)
				fmt.Println(errorPresc)
				if errorPresc := prescRequest.ParseTemplate(iterateTemplate+".html", data); errorPresc == nil {
					tableDrugs += prescRequest.body
					fmt.Println(prescRequest.body)
				}

				data.Name = ""
				data.Dosagedrug = ""
				data.Index = strconv.Itoa(index)
				data.Dosage = ""
				data.Pharmform = ""
				data.Package = ""
				data.Quantity = ""
				data.Lowest = ""
				data.Expiration = ""
				data.Instruction = ""
			}

			data.Name = convertToString(prescriptionContent[i][1])

			data.Dosage = convertToString(prescriptionContent[i][2])

			data.Quantity = convertToString(prescriptionContent[i][5])

			data.Lowest = convertToString(prescriptionContent[i][6])

			data.Expiration = convertToString(prescriptionContent[i][7])

			data.Instruction = convertToString(prescriptionContent[i][8])

			if convertToString(prescriptionContent[i][3]) == "forma_farmaceutica" {
				data.Pharmform = convertToString(prescriptionContent[i][4])

			} else if convertToString(prescriptionContent[i][3]) == "embalagem" {
				data.Package = convertToString(prescriptionContent[i][4])

			} else if convertToString(prescriptionContent[i][3]) == "qtd_embalagem" {
				data.Dosagedrug = convertToString(prescriptionContent[i][4])
			}

			requestId = prescId.(string)
			index = index + 1
		} else {
			if convertToString(prescriptionContent[i][3]) == "forma_farmaceutica" {
				data.Pharmform = convertToString(prescriptionContent[i][4])

			} else if convertToString(prescriptionContent[i][3]) == "embalagem" {
				data.Package = convertToString(prescriptionContent[i][4])

			} else if convertToString(prescriptionContent[i][3]) == "qtd_embalagem" {
				data.Dosagedrug = convertToString(prescriptionContent[i][4])
			}
		}

		if i == len(prescriptionContent)-1 {
			prescRequest := NewRequest([]string{""}, subject, "")
			errorPresc := prescRequest.ParseTemplate(iterateTemplate+".html", data)
			fmt.Println(errorPresc)
			if errorPresc := prescRequest.ParseTemplate(iterateTemplate+".html", data); errorPresc == nil {
				tableDrugs += prescRequest.body
				fmt.Println(prescRequest.body)
			}
		}
	}

	dispensationPin := ctx.GetInput("5_c_prescription_dispensation_pin").(string)
	optionPin := ctx.GetInput("5_d_prescription_option_pin").(string)
	expirationDate := ctx.GetInput("5_e_prescription_expiration_date").(string)
	prescriptionIdTransf := ctx.GetInput("5_a_prescription_id").(string)
	prescriptionIdBd := ctx.GetInput("5_b_prescription_id_db").(string)

	contact := ctx.GetInput("2_a_patient_contact").(string)
	from := fromName + " <" + emailFrom + ">"

	sampleMsg := fmt.Sprintf("From: %s\r\n", from)
	sampleMsg += fmt.Sprintf("To: %s\r\n", contact)
	sampleMsg += "Subject: " + subject + "\r\n"
	sampleMsg += "MIME-Version: 1.0\r\n"
	sampleMsg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	sampleMsg += "Content-Transfer-Encoding: 7bit\r\n"

	to := []string{""}

	if bcc == "" {
		to = []string{contact}
	} else {
		to = []string{contact, bcc}
	}

	templateData := struct {
		Number        string
		DismissalCode string
		RightCode     string
		Date          string
	}{
		Number:        prescriptionIdTransf,
		DismissalCode: dispensationPin,
		RightCode:     optionPin,
		Date:          expirationDate,
	}

	footer := ""
	fo := NewRequest([]string{contact}, subject, "")
	errory := fo.ParseTemplate(footerTemplate+".html", templateData)
	fmt.Println(errory)
	if errory := fo.ParseTemplate(footerTemplate+".html", templateData); errory == nil {
		footer = fo.body
	}

	r := NewRequest([]string{contact}, subject, "")
	error1 := r.ParseTemplate(contentTemplate+".html", templateData)
	fmt.Println(error1)
	if error1 := r.ParseTemplate(contentTemplate+".html", templateData); error1 == nil {

		sampleMsg += r.body
		sampleMsg += tableDrugs
		sampleMsg += footer

		if ssl != "true" {
			auth := smtp.PlainAuth("", emailAuth, password, server)
			err := smtp.SendMail(server+":"+port, auth, emailFrom, to, []byte(sampleMsg))
			if err != nil {
				fmt.Println(err)
				handleError(endpoint, prescriptionIdBd)
				email = ""
				success = false
			} else {
				output = sampleMsg
				success = true

			}
		} else {
			err := smtp.SendMail(server+":"+port, nil, emailFrom, to, []byte(sampleMsg))
			if err != nil {
				fmt.Println(err)
				handleError(endpoint, prescriptionIdBd)
				email = ""
				success = false
			} else {
				output = sampleMsg
				success = true
			}
		}
		log.Print("done.")
	}

	return output, success
}

func convertToString(text interface{}) string {
	text = *text.(*string)
	fText := text.(string)

	return fText
}

func createAppointment(ctx activity.Context) (email string, success bool) {
	//get input vars
	server := ctx.GetInput("1_b_smtp_server").(string)
	port := ctx.GetInput("1_c_smtp_port").(string)
	emailAuth := ctx.GetInput("1_e_smtp_auth_email").(string)
	fromName := ctx.GetInput("1_h_smtp_from_email").(string)
	ssl := ctx.GetInput("1_d_smtp_ssl").(string)
	bcc := ctx.GetInput("1_i_smtp_bcc_email").(string)
	password := ""
	emailFrom := emailAuth

	if ssl != "true" {
		password = ctx.GetInput("1_f_smtp_auth_password").(string)
		emailFrom = ctx.GetInput("1_h_smtp_from_email").(string)

	}

	appointment := ctx.GetInput("4_a_appointment_name").(string)
	date := ctx.GetInput("4_b_appointment_date").(string)

	clinic := ctx.GetInput("4_c_appointment_hospital").(string)
	meet := ctx.GetInput("4_d_appointment_meet").(string)
	subject := ctx.GetInput("1_l_subject").(string)
	status := ctx.GetInput("4_e_appointment_status").(string)
	appointmentId := ctx.GetInput("4_f_appointment_id").(string)
	endDate := ctx.GetInput("4_g_appointment_end_date").(string)
	appointmentIntId := ctx.GetInput("4_h_appointment_int_id").(string)

	contact := ctx.GetInput("2_a_patient_contact").(string)
	patient := ctx.GetInput("2_b_patient_name").(string)

	practitioner := ctx.GetInput("3_a_practitioner_name").(string)

	template := ctx.GetInput("4_i_appointment_template").(string)

	organizer := ctx.GetInput("4_j_ics_organizer").(string)
	prodid := ctx.GetInput("4_l_ics_prodid").(string)

	endpoint := ctx.GetInput("1_j_smtp_error_endpoint").(string)

	preparation := ctx.GetInput("4_m_appointment_preparation")

	templatePreparation := ctx.GetInput("4_n_appointment_template_preparation").(string)

	var preparationArray [][]interface{}

	if preparation != nil {
		preparationArray = preparation.([][]interface{})

	}

	method := "CANCEL"
	fstatus := "CANCELLED"
	transp := "TRANSPARENT"
	if status != "cancelled" {
		method = "PUBLISH"
		fstatus = "CONFIRMED"
		transp = "OPAQUE"
	}

	date1 := time.Now()
	fDate1 := date1.Format("20060102T150405Z")

	loc, err := time.LoadLocation("Europe/Lisbon")
	layout := "2006-01-02T15:04:05.000-0700"
	if err != nil {
		fmt.Println(err)
	}
	startDate, errd := time.Parse(layout, date)

	fEndDate, errd := time.Parse(layout, endDate)

	if errd != nil {
		fmt.Println(errd)
	}

	content := "BEGIN:VCALENDAR\r" +
		"METHOD:" + method + "\r" +
		"PRODID:" + prodid + "\r" +
		"VERSION:2.0\r" +
		"X-WR-TIMEZONE:Europe/Lisbon\r" +
		"BEGIN:VTIMEZONE\r" +
		"TZID:Europe/Lisbon\r" +
		"X-LIC-LOCATION:Europe/Lisbon\r" +
		"LAST-MODIFIED:20050809T050000Z\r" +
		"BEGIN:STANDARD\r" +
		"DTSTART:20071104T020000\r" +
		"TZOFFSETFROM:+0100\r" +
		"TZOFFSETTO:+0000\r" +
		"TZNAME:WET\r" +
		"END:STANDARD\r" +
		"BEGIN:DAYLIGHT\r" +
		"DTSTART:20070311T020000\r" +
		"TZOFFSETFROM:+0000\r" +
		"TZOFFSETTO:+0100\r" +
		"TZNAME:WEST\r" +
		"END:DAYLIGHT\r" +
		"END:VTIMEZONE\r" +
		"BEGIN:VEVENT\r" +
		"DTSTAMP:" + fDate1 + "\r" +
		"UID:" + appointmentId + "\r" +
		"SEQUENCE:0\r" +
		"ORGANIZER;" + organizer + "\r" +
		"DTSTART:" + startDate.Format("20060102T150405Z") + "\r" +
		"DTEND:" + fEndDate.Format("20060102T150405Z") + "\r" +
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

	var (
		serverAddr         = server
		portNumber         = port
		tos                = contact
		attachmentFilePath = filename1
		filename           = "invite.ics"
		delimeter          = "**=cuf689407924327"
	)

	from := fromName + " <" + emailFrom + ">"

	log.Println("Init create sample msg...")

	sampleMsg := fmt.Sprintf("From: %s\r\n", from)
	sampleMsg += fmt.Sprintf("To: %s\r\n", tos)
	sampleMsg += "Subject: " + subject + "\r\n"
	sampleMsg += "MIME-Version: 1.0\r\n"
	sampleMsg += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimeter)
	sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
	sampleMsg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	sampleMsg += "Content-Transfer-Encoding: 7bit\r\n"

	startDate = startDate.In(loc)
	fEndDate = fEndDate.In(loc)

	data := struct {
		PrepTitle string
		DescExam  string
		DescPrep  string
		Info      string
	}{
		PrepTitle: "",
		DescExam:  "",
		DescPrep:  "",
		Info:      "",
	}

	bucketFiles :=  ctx.GetInput("4_o_appointment_preparation_files")

	var linkBucketFiles string
	if bucketFiles != nil{
		linkBucketFiles = bucketFiles.(string)
	}


	var files []string

	if preparation != nil {

		for i := 0; i < len(preparationArray); i++ {


			if cast.ToString(preparationArray[i][0]) == "" {
				log.Println("build preparation files...")
				err := downloadFile(cast.ToString(preparationArray[i][4])+".pdf", linkBucketFiles+cast.ToString(preparationArray[i][3])+".pdf")
				fmt.Println(err)
				if err {
					files = append(files, cast.ToString(preparationArray[i][4])+".pdf")
				}
			} else {
				log.Println("build preparation...")

				if cast.ToString(preparationArray[i][2]) == "TITULO_PREPARACAO" {
					data.PrepTitle = cast.ToString(preparationArray[i][0])
				} else if cast.ToString(preparationArray[i][2]) == "DESCRICAO_PREPARACAO" {
					if data.DescPrep != "" {
						data.DescPrep += "\n"
					}
					data.DescPrep += cast.ToString(preparationArray[i][0])
				} else if cast.ToString(preparationArray[i][2]) == "INFORMACAO_ADICIONAL" {
					if data.Info != "" {
						data.Info += "\n"
					}
					data.Info += cast.ToString(preparationArray[i][0])
				} else if cast.ToString(preparationArray[i][2]) == "DESCRICAO_EXAME" {
					if data.DescExam != "" {
						data.DescExam += "\n"
					}
					data.DescExam += cast.ToString(preparationArray[i][0])
				}
			}
		}
	}

	isPreparation := false
	if data.PrepTitle != "" || data.DescExam != "" || data.Info != "" || data.DescPrep != "" {
		isPreparation = true
	}

	templateData := struct {
		Name          string
		Appointment   string
		Practitioner  string
		Date          string
		Hour          string
		Meet          string
		Hospital      string
		IsPreparation bool
	}{
		Name:          patient,
		Appointment:   appointment,
		Practitioner:  practitioner,
		Date:          strconv.Itoa(startDate.Day()) + "/" + strconv.Itoa(int(startDate.Month())),
		Hour:          handleHour(startDate.Hour()) + ":" + handleHour(startDate.Minute()),
		Meet:          meet,
		Hospital:      clinic,
		IsPreparation: isPreparation,
	}

	preparationText := ""

	if (len(preparationArray) > 0 || preparation != nil) && (data.PrepTitle != "" || data.DescPrep != "" || data.Info != "" || data.DescExam != "") {
		log.Println("Build preparation template...")
		if data.DescPrep != "" {
			prepRequest := NewRequest([]string{""}, subject, "")
			errorPrep := prepRequest.ParseTemplate(templatePreparation+".html", data)
			fmt.Println(errorPrep)
			if errorPrep := prepRequest.ParseTemplate(templatePreparation+".html", data); errorPrep == nil {
				preparationText += prepRequest.body
				fmt.Println(prepRequest.body)
			}
		}
	}

	log.Println("Build Footer...")
	footer := ""
	fo := NewRequest([]string{contact}, subject, "")
	errory := fo.ParseTemplate(template+"-footer.html", templateData)
	if errory != nil {
		fmt.Println(errory)
	}
	if errory := fo.ParseTemplate(template+"-footer.html", templateData); errory == nil {
		footer = fo.body
	}

	log.Println("Build content...")
	r := NewRequest([]string{contact}, subject, "")
	error1 := r.ParseTemplate(template+".html", templateData)
	if error1 := r.ParseTemplate(template+".html", templateData); error1 == nil {
		sampleMsg += r.body
		sampleMsg += preparationText
		sampleMsg += footer

		log.Println("Attach ics...")
		sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
		sampleMsg += "Content-Type: text/calendar; charset=\"utf-8\"\r\n"
		sampleMsg += "Content-Transfer-Encoding: base64\r\n"
		sampleMsg += "Content-Disposition: attachment;filename=\"" + filename + "\"\r\n"

		rawFile, fileErr := ioutil.ReadFile(attachmentFilePath)
		if fileErr != nil {
			log.Panic(fileErr)
		}

		sampleMsg += "\r\n" + base64.StdEncoding.EncodeToString(rawFile)

		fmt.Println(len(files))
		if len(files) > 0 {
			log.Println("Attach preparation files...")

			sampleMsg += fmt.Sprintf("\r\n--%s\r\n", delimeter)
			sampleMsg += "Content-Type: application/pdf; charset=\"utf-8\"\r\n"
			sampleMsg += "Content-Transfer-Encoding: base64\r\n"

			for k := 0; k < len(files); k++ {
				sampleMsg += "Content-Disposition: attachment;filename=\"" + files[k] + "\"\r\n"

				rawFile, fileErr := ioutil.ReadFile(files[k])
				if fileErr != nil {
					log.Panic(fileErr)
				}
				sampleMsg += "\r\n" + base64.StdEncoding.EncodeToString(rawFile)

				err1 := os.Remove(files[k]) // remove a single file
				if err1 != nil {
					fmt.Println(err1)
				}
			}
		}

		log.Println("Write content into client writter I/O")

		to := []string{tos, bcc}

		if bcc == "" {
			to = []string{contact}
		} else {
			to = []string{contact, bcc}
		}

		if ssl != "true" {
			log.Println("Send email...")
			auth := smtp.PlainAuth("", emailAuth, password, serverAddr)
			err := smtp.SendMail(serverAddr+":"+portNumber, auth, emailFrom, to, []byte(sampleMsg))
			if err != nil {
				fmt.Println(err)
				handleError(endpoint, appointmentIntId)
				success = false
				email = ""
			} else {
				email = sampleMsg
				success = true
			}
		} else {
			log.Println("Send email...")
			err := smtp.SendMail(serverAddr+":"+portNumber, nil, emailFrom, to, []byte(sampleMsg))
			if err != nil {
				fmt.Println(err)
				handleError(endpoint, appointmentIntId)
				success = false
				email = ""
			} else {
				email = sampleMsg
				success = true
			}
		}

		log.Print("done.")

		defer os.Remove(filename)

	}
	if error1 != nil {
		fmt.Println(error1)
	}

	return email, success
}

func CreateTempFile(serializer string) string {

	log.Println("Create ics...")

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
	if err1 == nil {
		fmt.Println(err1)
	}
	response, err := http.Post(endpoint+"/"+id, "application/json", bytes.NewBuffer(requestBody))
	if err == nil {
		fmt.Printf("The HTTP request failed with error %s\n", err)
	} else {
		data, _ := ioutil.ReadAll(response.Body)
		fmt.Println(string(data))
	}
	fmt.Println("Terminating retry update")
}

func handleHour(number int) (formatted string) {
	formatted = strconv.Itoa(number)
	if number == 0 {
		formatted = "00"
	} else {
		text := strconv.Itoa(number)
		if len(text) == 1 {
			formatted = "0" + text
		}
	}
	return formatted
}

func downloadFile(filepath string, url string) bool {
	// Get the data
	success := false
	resp, err := http.Get(url)
	if err != nil {
		log.Println(err)
	}
	defer resp.Body.Close()


	if resp.StatusCode == 200 {
		// Create the file
		out, err := os.Create(filepath)
		if err != nil {
			fmt.Println(err)
		}
		defer out.Close()

		// Write the body to file
		_, err = io.Copy(out, resp.Body)
		success = true
	}

	return success
}
