package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

type ContactMessage struct {
	ID          int
	FullName    string
	Gender      string
	FromDate    string
	ToDate      string
	PhoneNumber string
	Resume      string
	Email       string
	Message     string
}

var db *sql.DB

func main() {
	// Update these with your PostgreSQL connection details
	dbHost := "localhost"
	dbPort := 5432
	dbUser := "postgres"
	dbPassword := "Haripriya@2001"
	dbName := "xenonstack"

	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPassword, dbName)

	var err error
	db, err = sql.Open("postgres", connectionString)
	if err != nil {
		fmt.Println("Error opening database:", err)
		return
	}
	defer db.Close()

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS messages (
			id INTEGER PRIMARY KEY,
			full_name TEXT,
			gender TEXT,
			from_date TEXT,
			to_date TEXT,
			phone_number TEXT,
			resume TEXT,
			email TEXT,
			message TEXT
		);
    `)

	if err != nil {
		fmt.Println("Error creating table:", err)
		return
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	http.HandleFunc("/", contactFormHandler)
	http.HandleFunc("/submit", submitHandler)
	http.HandleFunc("/messages", getMessagesHandler)
	http.HandleFunc("/message", getMessageByIDHandler)

	fmt.Printf("Server started on http://localhost:%s\n", port)
	http.ListenAndServe(":"+port, nil)
}

func contactFormHandler(w http.ResponseWriter, r *http.Request) {
	tmpl := `
	<!DOCTYPE html>

<html lang="en" dir="ltr">
  <head>
    <meta charset="UTF-8">
   <style>
    @import url('https://fonts.googleapis.com/css2?family=Poppins:wght@200;300;400;500;600;700&display=swap');
	*{
		margin: 0;
		padding: 0;
		box-sizing: border-box;
		font-family: 'Poppins',sans-serif;
	  }
	  body {
		height: 100vh;
		display: flex;
		justify-content: center;
		align-items: center;
		padding: 10px;
		background: linear-gradient(135deg, #edddd8, #0c9b4d);
		position: relative; /* Add this to the body */
	  }
	  
	  
		  /* Style for the welcome text */
		  .welcome-text {
			position: absolute;
			left: -300px;
			top: 50%;
			transform: translateY(-50%);
			font-family: 'Pacifico', cursive;
			font-size: 50px;
			font-weight: 600;
			color: white;
			opacity: 0;
			transition: left 0.5s ease-in-out, opacity 0.5s ease-in-out;
			animation: glow 2s infinite alternate; /* Adding a glow animation */
		  }
	  
		  /* Add a sparkle effect using pseudo-elements */
		  .welcome-text::before,
		  .welcome-text::after {
			content: '';
			position: absolute;
			width: 10px;
			height: 10px;
			background-color: white;
			border-radius: 50%;
			animation: sparkle 1.5s infinite;
		  }
	  
		  .welcome-text::before {
			top: -5px;
			left: -10px;
		  }
	  
		  .welcome-text::after {
			bottom: -5px;
			right: -10px;
		  }
	  
		  /* When hovering over the body, move the text to the desired position and fade it in */
		  body:hover .welcome-text {
			left: 20px;
			opacity: 1;
		  }
	  
		 
		  @keyframes glow-welcome {
			from {
			  text-shadow: 0 0 10px rgba(255, 255, 255, 0.5);
			}
			to {
			  text-shadow: 0 0 20px rgba(255, 255, 255, 1);
			}
		  }
	  
		  /* Sparkle animation */
		  @keyframes sparkle {
			0%, 100% {
			  transform: scale(1);
			  opacity: 1;
			}
			50% {
			  transform: scale(1.5);
			  opacity: 0;
			}
		  }
	  
	 
		  .container {
			max-width: 600px;
			width: 100%;
			background-color: rgba(255, 255, 255, 0.2);
			padding: 20px 25px;
			border-radius: 5px;
			box-shadow: 0 5px 10px rgba(0, 0, 0, 0.15);
			backdrop-filter: blur(10px);
			animation: glow 2s infinite alternate; /* Adding a glow animation */
		  
			/* Positioning styles */
			position: absolute;
			top: 50%;
			right: 15px;
			transform: translateY(-50%);
		  }
		  
		  /* Glow animation for the container */
		  @keyframes glow-container {
		 from {
		 box-shadow: 0 0 10px rgba(255, 255, 255, 0.5);
		 }
		 to {
		 box-shadow: 0 0 20px rgba(255, 255, 255, 1);
		}
		  }
			.container .title{
			  font-size: 25px;
			  font-weight: 500;
			  position: relative;
			}
	  .container .title::before{
		content: "";
		position: absolute;
		left: 0;
		bottom: 0;
		height: 3px;
		width: 100px;
		border-radius: 5px;
		background: linear-gradient(135deg, #e2ed11, #f05908);
	  }
	  .content form .user-details{
		display: flex;
		flex-wrap: wrap;
		justify-content: space-between;
		margin: 20px 0 12px 0;
	  }
	  
	  form .user-details .input-box{
		margin-bottom: 15px;
		width: calc(100% / 2 - 20px);
	  }
	  
	  form .input-box span.details{
		display: block;
		font-weight: 500;
		margin-bottom: 5px;
	  }
	  .user-details .input-box input{
		height: 45px;
		width: 100%;
		outline: none;
		font-size: 16px;
		border-radius: 5px;
		padding-left: 15px;
		border: 1px solid #ccc;
		border-bottom-width: 2px;
		transition: all 0.3s ease;
	  }
	  
	  .user-details .input-box input:focus,
	  .user-details .input-box input:valid{
		border-color: #0a0a0a;
	  }
	  /* Style for gender dropdown */
.gender-details {
  margin-top: 20px;
  font-weight: 500;
  color: #3a084e;
}

.gender-details label {
  display: block;
  margin-bottom: 5px;
}

.gender-details select {
  width: 100%;
  padding: 8px;
  border-radius: 5px;
  border: 1px solid #ccc;
  background-color: #f7f7f7;
  font-size: 16px;
  outline: none;
  color: black;
  transition: border-color 0.3s ease;
}

.gender-details select:focus {
  border-color: #040404;
}

	  
.button {
	height: 30px; /* Increased height */
	margin: 35px 0;
	display: flex;
	justify-content: center;
	align-items: center;
  }
  
  .button input {
	height: 100%;
	width: 100%; /* You can adjust the width as needed */
	border-radius: 8px; /* Rounded corners */
	border: none;
	color: #fff;
	font-size: 20px;
	font-weight: 600;
	letter-spacing: 1px;
	cursor: pointer;
	transition: background 0.3s ease;
	background: linear-gradient(135deg, #2b1858, #4b0a54);
	outline: none;
  }
  
  .button input:hover {
	background: linear-gradient(-135deg, #d3fa5c, #08ff67);
  }

 @media(max-width: 584px){
 .container{
  max-width: 100%;
}
form .user-details .input-box{
    margin-bottom: 15px;
    width: 200%;
  }
  form .category{
    width: 100%;
  }
  .content form .user-details{
    max-height: 300px;
    overflow-y: scroll;
  }
  .user-details::-webkit-scrollbar{
    width: 5px;
  }
  }
  @media(max-width: 459px){
  .container .content .category{
    flex-direction: column;
  }
}
   </style>

  

   <meta charset="UTF-8">
   <title> Responsive Registration Form</title>
   
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
  </head>
<body>
<div class="welcome-text">Welcome To Xenonstack</div>
 <div class="container">
   <div class="title">Employee Details</div>
   <div class="content">
    <form action="/submit" method="post" enctype="multipart/form-data">
       <div class="user-details">
         <div class="input-box">
            <span class="details">Full Name</span>
            <input type="text" placeholder="Enter your name" name="full_name" required>

          </div>
          <div class="input-box">
            <span class="username">Email</span>
            <input type="text" placeholder="Enter your business email" name="email" required>

          </div>
          <div class="input-box">
            <span class="details">Phone Number</span>
            <input type="tel" id="phone_number" name="phone_number" pattern="[0-9]+" required>
          </div>
          
          <div class="input-box">
            <span class="details">FROM Date</span>
            <label for="date_range"></label>
            <input type="date" id="from_date" name="from_date" required>
          </div>
          <div class="input-box">
            <span class="details">To Date</span>
            <label for="date_range">Date: To Date</label><br>
            <input type="date" id="to_date" name="to_date" required>
          </div>
          <div class="input-box">
            <span class="details">Resume</span>
            <label for="resume">Resume (PDF/PNG, max 5MB):</label><br>
	          <input type="file" id="resume" name="resume" accept=".pdf,.png" required><br>
          </div>
        </div>
        <div class="gender-details">
        <label for="gender" style="color: black;">Gender:</label><br>
          <select id="gender" name="gender" required>

		         <option value="Male">Male</option>
		        <option value="Female">Female</option>
		        <option value="Others">Others</option>
	          </select><br>
          </div>
        </div>
        <br><br>
        <div class="input-box">
          <span class="details"></span>
          <label for="message">Message:</label><br>
          <textarea id="message" name="message" rows="4" style="width: 100%; height: 50px;" required></textarea><br>
        </div>
        

        <div class="button">
          <input type="submit" value="Submit">
        </div>
      </form>
    </div>

  </div>

</body>
</html>
	`

	tmplParsed := template.Must(template.New("contactForm").Parse(tmpl))
	tmplParsed.Execute(w, nil)
}

func submitHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	err := r.ParseMultipartForm(5 * 1024 * 1024) // Max file size 5MB
	if err != nil {
		fmt.Println("Error parsing form data:", err)
		http.Error(w, "Error parsing form data", http.StatusBadRequest)
		return
	}

	contact := ContactMessage{
		FullName:    strings.TrimSpace(r.FormValue("full_name")),
		Gender:      r.FormValue("gender"),
		FromDate:    r.FormValue("from_date"),
		ToDate:      r.FormValue("to_date"),
		PhoneNumber: r.FormValue("phone_number"),
		Email:       r.FormValue("email"),

		Message: r.FormValue("message"),
	}

	if !isValidPhoneNumber(contact.PhoneNumber) {
		http.Error(w, "Please enter a valid Indian phone number", http.StatusBadRequest)
		return
	}

	if !isValidEmail(contact.Email) {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	// Retrieve the uploaded resume file
	resumeFile, resumeHeader, err := r.FormFile("resume")
	if err == nil {
		defer resumeFile.Close() // Close the file when done

		// Resume validation
		if !isValidResume(resumeHeader) {
			http.Error(w, "Invalid resume file. Only PDF and PNG files up to 5MB are allowed.", http.StatusBadRequest)
			return
		}

		// Generate a unique filename for the resume
		resumeFileName := fmt.Sprintf("resume_%d%s", time.Now().Unix(), filepath.Ext(resumeHeader.Filename))

		// Save the resume file to the 'resumes' directory
		resumeFilePath := filepath.Join("resumes", resumeFileName)
		outFile, err := os.Create(resumeFilePath)
		if err != nil {
			http.Error(w, "Error saving resume file", http.StatusInternalServerError)
			return
		}
		defer outFile.Close()

		_, err = io.Copy(outFile, resumeFile)
		if err != nil {
			http.Error(w, "Error saving resume file", http.StatusInternalServerError)
			return
		}

		// Assign the resume file name to the contact.Resume field
		contact.Resume = resumeFileName
	}

	fromDateParsed, err := time.Parse("2006-01-02", contact.FromDate)
	if err != nil {
		http.Error(w, "Invalid From Date format", http.StatusBadRequest)
		return
	}

	toDateParsed, err := time.Parse("2006-01-02", contact.ToDate)
	if err != nil {
		http.Error(w, "Invalid To Date format", http.StatusBadRequest)
		return
	}

	if toDateParsed.Before(fromDateParsed) || toDateParsed.Equal(fromDateParsed) {
		http.Error(w, "To Date should be greater than From Date", http.StatusBadRequest)
		return
	}

	if len(contact.FullName) >= 30 {
		http.Error(w, "Full Name should be less than 30 characters", http.StatusBadRequest)
		return
	}

	sqlStatement := "INSERT INTO messages (full_name, gender, from_date, to_date, phone_number, resume, email, message) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)"

	_, err = db.Exec(sqlStatement, contact.FullName, contact.Gender, contact.FromDate, contact.ToDate, contact.PhoneNumber, contact.Resume, contact.Email, contact.Message)
	if err != nil {
		fmt.Println("Error storing data in the database:", err)
		http.Error(w, "Error storing data in the database", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "Message submitted successfully!")
}

func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := db.Query("SELECT id, full_name, gender, from_date, to_date, phone_number, email, message FROM messages")
	if err != nil {
		http.Error(w, "Error retrieving data from the database", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var messages []ContactMessage
	for rows.Next() {
		var msg ContactMessage
		err := rows.Scan(&msg.ID, &msg.FullName, &msg.Gender, &msg.FromDate, &msg.ToDate, &msg.PhoneNumber, &msg.Email, &msg.Message)
		if err != nil {
			http.Error(w, "Error scanning database rows", http.StatusInternalServerError)
			return
		}
		messages = append(messages, msg)
	}

	fmt.Fprintf(w, "<h1>Contact Messages</h1>")
	for _, msg := range messages {
		fmt.Fprintf(w, "<p><strong>Full Name:</strong> %s<br><strong>Gender:</strong> %s<br><strong>From Date:</strong> %s<br><strong>To Date:</strong> %s<br><strong>Phone Number:</strong> %s<br><strong>Email:</strong> %s<br><strong>Message:</strong> %s</p><hr>", msg.FullName, msg.Gender, msg.FromDate, msg.ToDate, msg.PhoneNumber, msg.Email, msg.Message)
	}
}

func getMessageByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "ID parameter is missing", http.StatusBadRequest)
		return
	}

	var msg ContactMessage
	err := db.QueryRow("SELECT id, full_name, gender, from_date, to_date, phone_number, email, resume, message FROM messages WHERE id = $1", id).Scan(&msg.ID, &msg.FullName, &msg.Gender, &msg.FromDate, &msg.ToDate, &msg.PhoneNumber, &msg.Email, &msg.Resume, &msg.Message)
	if err != nil {
		http.Error(w, "Error retrieving data from the database", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "<h1>Contact Message with ID %s</h1>", id)
	fmt.Fprintf(w, "<p><strong>ID:</strong> %d<br><strong>Full Name:</strong> %s<br><strong>Gender:</strong> %s<br><strong>From Date:</strong> %s<br><strong>To Date:</strong> %s<br><strong>Phone Number:</strong> %s<br><strong>Email:</strong> %s<br><strong>Resume:</strong> %s<br><strong>Message:</strong> %s</p><hr>", msg.ID, msg.FullName, msg.Gender, msg.FromDate, msg.ToDate, msg.PhoneNumber, msg.Email, msg.Resume, msg.Message)
}

func isValidPhoneNumber(number string) bool {
	// The regular expression to match Indian phone numbers (10 digits starting with 7, 8, or 9)
	match, _ := regexp.MatchString(`^(?:\+91|0?91|\d{0,2}-?)?[7-9]\d{9}$`, number)
	return match
}

func isValidEmail(email string) bool {
	emailPattern := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.(com|net|org|edu|biz|info|io|your-extension-here)$`

	match, _ := regexp.MatchString(emailPattern, email)
	return match
}

func isValidResume(resumeHeader *multipart.FileHeader) bool {
	if resumeHeader != nil {
		if resumeHeader.Size <= 5*1024*1024 && (resumeHeader.Header.Get("Content-Type") == "application/pdf" || resumeHeader.Header.Get("Content-Type") == "image/png") {
			return true
		}
	}
	return false
}
