# WhatsMyIP

This is a simple Go web application that presents a web page showing the client's source IP and related headers such as X-Forwarded-* and host header.

## How to Run

1. Make sure you have Go installed on your machine.
2. Clone this repository.
3. Navigate to the project directory.
4. Run the following command to start the web server:

   ```sh
   go run main.go
   ```

5. Open your web browser and go to `http://localhost:8080`.

## Example Output

When you visit the web page, you will see something like this:

```
Client IP: [your IP address]
Headers:
X-Forwarded-For: [value]
X-Forwarded-Host: [value]
X-Forwarded-Proto: [value]
Host: [value]
```
