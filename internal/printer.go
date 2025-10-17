package internal

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type PrinterServer struct {
	server *echo.Echo
	prefs  *PrefsStore
}

type CommandRequest struct {
	Command string `json:"command"`
}

type PrintRequest struct {
	Data []byte `json:"data"`
}

func NewPrinterServer(prefs *PrefsStore) *PrinterServer {
	e := echo.New()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"http://localhost:5174",
			"https://kitchen.otterorder.com",
		},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodOptions,
		},
		AllowHeaders: []string{
			"Content-Type",
			"Authorization",
		},
	}))

	ps := &PrinterServer{
		server: e,
		prefs:  prefs,
	}

	ps.server.POST("/", ps.handlePrint)

	return ps
}

func (ps *PrinterServer) PrintTestPage() error {
	fmt.Println("Starting print test...")

	p, err := ps.prefs.GetPreferences()
	if err != nil {
		return fmt.Errorf("failed to get preferences: %w", err)
	}

	if p.PrinterIP == "" || p.PrinterPort == "" {
		return fmt.Errorf("printer IP or port not set")
	}

	fmt.Printf("Connecting to printer at %s:%s\n", p.PrinterIP, p.PrinterPort)

	addr := net.JoinHostPort(p.PrinterIP, p.PrinterPort)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to connect to printer: %w", err)
	}
	defer conn.Close()

	fmt.Println("Connected successfully!")

	const ESC = "\x1B"
	const GS = "\x1D"

	testPage := ""
	testPage += ESC + "@"
	testPage += ESC + "a" + "\x01"
	testPage += ESC + "!" + "\x38"
	testPage += "OTTER ORDER\n"
	testPage += ESC + "!" + "\x00"
	testPage += ESC + "a" + "\x00"
	testPage += "Printer Bridge Test Page\n"
	testPage += "============================\n"
	testPage += "Connection successful!\n"
	testPage += fmt.Sprintf("IP:   %s\n", p.PrinterIP)
	testPage += fmt.Sprintf("Port: %s\n", p.PrinterPort)
	testPage += fmt.Sprintf("Time: %s\n", time.Now().Format("2006-01-02 15:04:05"))
	testPage += "============================\n\n"

	testPage += ESC + "a" + "\x01"
	testPage += "✅ PRINT OK ✅\n\n\n\n\n\n\n\n\n\n"
	testPage += GS + "V" + "\x01"

	fmt.Printf("Sending %d bytes to printer...\n", len(testPage))

	if _, err := conn.Write([]byte(testPage)); err != nil {
		return fmt.Errorf("failed to write to printer: %w", err)
	}

	fmt.Println("Print job sent successfully!")
	return nil
}

func (ps *PrinterServer) handlePrint(c echo.Context) error {
	var req PrintRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request format",
		})
	}

	p, err := ps.prefs.GetPreferences()

	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to get preferences: %v", err),
		})
	}

	err = ps.sendRawDataToSocket(p.PrinterIP, p.PrinterPort, req.Data)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("Failed to send data: %v", err),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"status": "Data sent successfully",
	})
}

func (ps *PrinterServer) sendRawDataToSocket(ipAddress, port string, data []byte) error {
	address := net.JoinHostPort(ipAddress, port)

	conn, err := net.DialTimeout("tcp", address, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to connect to %s: %w", address, err)
	}
	defer conn.Close()

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	_, err = conn.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data: %w", err)
	}

	return nil
}

func (ps *PrinterServer) StartWithContext(ctx context.Context) error {
	fmt.Println("Starting server on port 3838...")

	go func() {
		<-ctx.Done()
		fmt.Println("Shutting down server...")
		ps.server.Shutdown(context.Background())
	}()

	return ps.server.Start(":3838")
}
