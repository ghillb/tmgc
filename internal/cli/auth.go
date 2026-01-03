package cli

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"image/png"
	"io"
	"os"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/auth/qrlogin"
	"github.com/gotd/td/tg"
	"github.com/gotd/td/tgerr"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"rsc.io/qr"

	"github.com/ghillb/tmgc/internal/config"
	"github.com/ghillb/tmgc/internal/tgclient"
	"github.com/ghillb/tmgc/internal/types"
)

func newAuthCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authenticate with Telegram",
	}

	cmd.AddCommand(newAuthLoginCmd())
	cmd.AddCommand(newAuthStatusCmd())
	cmd.AddCommand(newAuthLogoutCmd())
	cmd.AddCommand(newAuthConfigCmd())

	return cmd
}

func newAuthLoginCmd() *cobra.Command {
	var (
		method   string
		phone    string
		apiID    int
		apiHash  string
		botToken string
		qrFile   string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login using QR or phone code",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}

			cfg := *rt.Config
			if apiID != 0 {
				cfg.APIID = apiID
			}
			if apiHash != "" {
				cfg.APIHash = apiHash
			}
			if cfg.APIID == 0 || cfg.APIHash == "" {
				return errors.New("missing API credentials: provide --api-id and --api-hash or set TMGC_API_ID/TMGC_API_HASH")
			}

			if apiID != 0 || apiHash != "" {
				if err := config.Save(rt.Paths.ConfigPath, cfg); err != nil {
					return err
				}
				*rt.Config = cfg
			}

			// Auth flows are interactive; don't impose a hard timeout.
			factory := tgclient.NewFactory(cfg, rt.Paths, rt.Printer, 0)
			return factory.Run(cmd.Context(), false, func(ctx context.Context, b *tgclient.Bundle) error {
				status, err := b.Client.Auth().Status(ctx)
				if err != nil {
					return err
				}
				if status.Authorized {
					return printAuthStatus(ctx, rt, b)
				}

				if botToken != "" {
					if _, err := b.Client.Auth().Bot(ctx, botToken); err != nil {
						return err
					}
					return printAuthStatus(ctx, rt, b)
				}

				switch method {
				case "qr", "":
					if err := loginQR(ctx, rt, b, qrRenderOptions{
						filePath: qrFile,
					}); err != nil {
						cleanupLoginFailure(ctx, rt, b)
						return err
					}
					return nil
				case "code":
					if phone == "" {
						return errors.New("--phone is required for code login")
					}
					if err := loginCode(ctx, rt, b, phone); err != nil {
						cleanupLoginFailure(ctx, rt, b)
						return err
					}
					return nil
				default:
					return fmt.Errorf("unknown login method: %s", method)
				}
			})
		},
	}

	cmd.Flags().StringVar(&method, "method", "qr", "login method: qr or code")
	cmd.Flags().StringVar(&phone, "phone", "", "phone number for code login")
	cmd.Flags().IntVar(&apiID, "api-id", 0, "telegram api id")
	cmd.Flags().StringVar(&apiHash, "api-hash", "", "telegram api hash")
	cmd.Flags().StringVar(&botToken, "bot-token", "", "bot token (optional)")
	cmd.Flags().StringVar(&qrFile, "qr-file", "", "write QR to PNG file (optional)")

	return cmd
}

type qrRenderOptions struct {
	filePath string
}

func loginQR(ctx context.Context, rt *Runtime, b *tgclient.Bundle, opts qrRenderOptions) error {
	loggedIn := qrlogin.OnLoginToken(b.Dispatcher)
	qr := b.Client.QR()
	var lastURL string
	var headerPrinted bool

	show := func(ctx context.Context, token qrlogin.Token) error {
		url := token.URL()
		if url == lastURL {
			return nil
		}
		lastURL = url
		if !headerPrinted {
			rt.Printer.Logln("Scan this QR in Telegram (Settings -> Devices -> Link Desktop Device):")
			headerPrinted = true
		} else {
			rt.Printer.Logln("QR rotated:")
		}
		if err := printASCIIQR(rt, url, opts); err != nil {
			rt.Printer.Logln(url)
		}
		return nil
	}

	_, err := qr.Auth(ctx, loggedIn, show)
	if err != nil {
		if tgerr.Is(err, "SESSION_PASSWORD_NEEDED") {
			if err := handlePassword(ctx, rt, b); err != nil {
				return err
			}
			return printAuthStatus(ctx, rt, b)
		}
		return err
	}
	return printAuthStatus(ctx, rt, b)
}

func loginCode(ctx context.Context, rt *Runtime, b *tgclient.Bundle, phone string) error {
	reader := bufio.NewReader(os.Stdin)
	codeAsk := func(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
		_ = ctx
		_ = sentCode
		rt.Printer.Logf("Enter code for %s: ", phone)
		code, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(code), nil
	}
	passAsk := func(ctx context.Context) (string, error) {
		_ = ctx
		return promptPassword(rt, reader)
	}

	authenticator := &loginAuthenticator{
		phone:    phone,
		codeAuth: auth.CodeAuthenticatorFunc(codeAsk),
		password: passAsk,
	}

	flow := auth.NewFlow(authenticator, auth.SendCodeOptions{})
	if err := flow.Run(ctx, b.Client.Auth()); err != nil {
		return err
	}
	return printAuthStatus(ctx, rt, b)
}

func handlePassword(ctx context.Context, rt *Runtime, b *tgclient.Bundle) error {
	reader := bufio.NewReader(os.Stdin)
	pass, err := promptPassword(rt, reader)
	if err != nil {
		return err
	}
	_, err = b.Client.Auth().Password(ctx, pass)
	return err
}

func promptPassword(rt *Runtime, reader *bufio.Reader) (string, error) {
	rt.Printer.Logf("Enter 2FA password: ")
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		pass, err := readPasswordMasked(fd, rt.Printer.Err)
		if err != nil {
			return "", err
		}
		return strings.TrimSpace(pass), nil
	}

	pass, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(pass), nil
}

func readPasswordMasked(fd int, out io.Writer) (string, error) {
	state, err := term.MakeRaw(fd)
	if err != nil {
		return "", err
	}
	defer term.Restore(fd, state)

	var buf bytes.Buffer
	reader := bufio.NewReader(os.Stdin)

	for {
		r, _, err := reader.ReadRune()
		if err != nil {
			return "", err
		}

		switch r {
		case '\r', '\n':
			fmt.Fprintln(out)
			return buf.String(), nil
		case 3: // Ctrl+C
			return "", errors.New("interrupted")
		case 4: // Ctrl+D
			return "", io.EOF
		case 127, 8: // Backspace/Delete
			if buf.Len() > 0 {
				// Remove last rune from buffer.
				b := buf.Bytes()
				_, size := utf8.DecodeLastRune(b)
				if size > 0 {
					buf.Truncate(len(b) - size)
					fmt.Fprint(out, "\b \b")
				}
			}
		default:
			buf.WriteRune(r)
			fmt.Fprint(out, "*")
		}
	}
}

type loginAuthenticator struct {
	phone    string
	codeAuth auth.CodeAuthenticator
	password func(ctx context.Context) (string, error)
}

func (l *loginAuthenticator) Phone(ctx context.Context) (string, error) {
	return l.phone, nil
}

func (l *loginAuthenticator) Password(ctx context.Context) (string, error) {
	if l.password == nil {
		return "", auth.ErrPasswordNotProvided
	}
	return l.password(ctx)
}

func (l *loginAuthenticator) AcceptTermsOfService(ctx context.Context, tos tg.HelpTermsOfService) error {
	_ = ctx
	_ = tos
	return errors.New("sign-up required; please complete sign-up in an official Telegram app")
}

func (l *loginAuthenticator) SignUp(ctx context.Context) (auth.UserInfo, error) {
	_ = ctx
	return auth.UserInfo{}, errors.New("sign-up required; please complete sign-up in an official Telegram app")
}

func (l *loginAuthenticator) Code(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	return l.codeAuth.Code(ctx, sentCode)
}

func newAuthStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show auth status",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}
			factory := tgclient.NewFactory(*rt.Config, rt.Paths, rt.Printer, rt.Timeout)
			return factory.Run(cmd.Context(), false, func(ctx context.Context, b *tgclient.Bundle) error {
				return printAuthStatus(ctx, rt, b)
			})
		},
	}
	return cmd
}

func newAuthLogoutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout",
		Short: "Log out and clear local session",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}
			factory := tgclient.NewFactory(*rt.Config, rt.Paths, rt.Printer, rt.Timeout)
			return factory.Run(cmd.Context(), false, func(ctx context.Context, b *tgclient.Bundle) error {
				_, _ = b.Client.API().AuthLogOut(ctx)
				tgclient.ClearSession(*rt.Config, rt.Paths, rt.Printer)
				return nil
			})
		},
	}
	return cmd
}

func printAuthStatus(ctx context.Context, rt *Runtime, b *tgclient.Bundle) error {
	status, err := b.Client.Auth().Status(ctx)
	if err != nil {
		return err
	}

	result := types.AuthStatus{Authorized: status.Authorized}
	if status.Authorized {
		user, err := b.Client.Self(ctx)
		if err != nil {
			return err
		}
		result.UserID = user.ID
		result.Username = user.Username
		result.Phone = user.Phone
		result.IsBot = user.Bot
	}

	switch rt.Printer.Mode {
	case "json":
		return rt.Printer.JSON(result)
	case "plain":
		line := fmt.Sprintf("%t\t%d\t%s\t%s\t%t", result.Authorized, result.UserID, result.Username, result.Phone, result.IsBot)
		rt.Printer.Plain([]string{line})
	default:
		if result.Authorized {
			fmt.Fprintln(rt.Printer.Out)
			rt.Printer.Table([][]string{{"AUTHORIZED", "USER", "USERNAME", "PHONE", "BOT"}, {
				strconv.FormatBool(result.Authorized),
				strconv.FormatInt(result.UserID, 10),
				result.Username,
				result.Phone,
				strconv.FormatBool(result.IsBot),
			}})
		} else {
			rt.Printer.Logln("Not authorized. Run `tmgc auth login`.")
		}
	}
	return nil
}

func cleanupLoginFailure(ctx context.Context, rt *Runtime, b *tgclient.Bundle) {
	_, _ = b.Client.API().AuthLogOut(ctx)
	tgclient.ClearSession(*rt.Config, rt.Paths, rt.Printer)
}

func newAuthConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage API credentials",
	}

	cmd.AddCommand(newAuthConfigSetCmd())
	cmd.AddCommand(newAuthConfigShowCmd())

	return cmd
}

func newAuthConfigSetCmd() *cobra.Command {
	var (
		apiID        int
		apiHash      string
		sessionStore string
	)

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set API credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}

			if apiID == 0 && apiHash == "" && sessionStore == "" {
				return errors.New("provide --api-id, --api-hash, and/or --session-store")
			}

			cfg := *rt.Config
			if apiID != 0 {
				cfg.APIID = apiID
			}
			if apiHash != "" {
				cfg.APIHash = apiHash
			}
			if sessionStore != "" {
				normalized, err := normalizeSessionStoreInput(sessionStore)
				if err != nil {
					return err
				}
				cfg.SessionStore = normalized
			}

			if err := config.Save(rt.Paths.ConfigPath, cfg); err != nil {
				return err
			}
			*rt.Config = cfg

			switch rt.Printer.Mode {
			case "json":
				return rt.Printer.JSON(cfg)
			case "plain":
				line := fmt.Sprintf("%d\t%s\t%s", cfg.APIID, cfg.APIHash, displaySessionStore(cfg.SessionStore))
				rt.Printer.Plain([]string{line})
			default:
				rt.Printer.Table([][]string{{"API_ID", "API_HASH", "SESSION_STORE"}, {
					strconv.Itoa(cfg.APIID),
					maskHash(cfg.APIHash),
					displaySessionStore(cfg.SessionStore),
				}})
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&apiID, "api-id", 0, "telegram api id")
	cmd.Flags().StringVar(&apiHash, "api-hash", "", "telegram api hash")
	cmd.Flags().StringVar(&sessionStore, "session-store", "", "session storage: keyring or file")

	return cmd
}

func newAuthConfigShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show current API credentials",
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := runtimeFrom(cmd.Context())
			if err != nil {
				return err
			}

			cfg := *rt.Config
			switch rt.Printer.Mode {
			case "json":
				return rt.Printer.JSON(cfg)
			case "plain":
				line := fmt.Sprintf("%d\t%s\t%s", cfg.APIID, cfg.APIHash, displaySessionStore(cfg.SessionStore))
				rt.Printer.Plain([]string{line})
			default:
				rt.Printer.Table([][]string{{"API_ID", "API_HASH", "SESSION_STORE"}, {
					strconv.Itoa(cfg.APIID),
					maskHash(cfg.APIHash),
					displaySessionStore(cfg.SessionStore),
				}})
			}
			return nil
		},
	}
	return cmd
}

func maskHash(hash string) string {
	if hash == "" {
		return ""
	}
	if len(hash) <= 6 {
		return "******"
	}
	return hash[:2] + strings.Repeat("*", len(hash)-4) + hash[len(hash)-2:]
}

func normalizeSessionStoreInput(store string) (string, error) {
	store = strings.TrimSpace(strings.ToLower(store))
	switch store {
	case "keyring", "file":
		return store, nil
	default:
		return "", fmt.Errorf("invalid session store %q (use keyring or file)", store)
	}
}

func displaySessionStore(store string) string {
	if strings.TrimSpace(store) == "" {
		return "default"
	}
	return store
}

func printASCIIQR(rt *Runtime, url string, opts qrRenderOptions) error {
	code, err := qr.Encode(url, qr.M)
	if err != nil {
		return err
	}

	if opts.filePath != "" {
		if err := writeQRPNG(opts.filePath, code); err != nil {
			rt.Printer.Logf("Failed to write QR PNG: %v\n", err)
		} else {
			rt.Printer.Logln("Saved QR PNG to:", opts.filePath)
		}
	}

	lines := renderASCIIQR(code)
	for _, line := range lines {
		rt.Printer.Logln(line)
	}
	return nil
}

func renderASCIIQR(code *qr.Code) []string {
	size := code.Size
	quiet := 4
	edge := size + quiet*2
	lines := make([]string, 0, (edge+1)/2)

	for y := -quiet; y < size+quiet; y += 2 {
		var b strings.Builder
		b.Grow(edge)
		for x := -quiet; x < size+quiet; x++ {
			top := false
			bottom := false
			if x >= 0 && x < size && y >= 0 && y < size {
				top = code.Black(x, y)
			}
			if x >= 0 && x < size && (y+1) >= 0 && (y+1) < size {
				bottom = code.Black(x, y+1)
			}

			switch {
			case top && bottom:
				b.WriteRune('█')
			case top && !bottom:
				b.WriteRune('▀')
			case !top && bottom:
				b.WriteRune('▄')
			default:
				b.WriteRune(' ')
			}
		}
		lines = append(lines, b.String())
	}
	return lines
}

func writeQRPNG(path string, code *qr.Code) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, code.Image())
}
