package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gabrandalisse/rant/internal/rant"
	"github.com/gabrandalisse/rant/internal/web"
	"github.com/spf13/cobra"
)

var dayNames = [7]string{"Domingo", "Lunes", "Martes", "Miércoles", "Jueves", "Viernes", "Sábado"}

func Execute(svc *rant.Service) error {
	var tags []string
	var severity int

	rootCmd := &cobra.Command{
		Use:           "rant [mensaje]",
		Short:         "Descargá bronca laboral de forma cifrada y local",
		Args:          cobra.MaximumNArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			var content string
			var err error

			if len(args) == 1 {
				content = args[0]
			} else {
				content, err = captureFromEditor()
				if err != nil {
					return err
				}
			}

			created, err := svc.Record(cmd.Context(), rant.RecordInput{
				Content:  content,
				Severity: severity,
				Tags:     tags,
			})
			if err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Registrado [%s] severidad %d/5\n", created.ID[:8], created.Severity)
			return nil
		},
	}

	rootCmd.CompletionOptions.DisableDefaultCmd = true
	rootCmd.Flags().StringSliceVarP(&tags, "tags", "t", nil, "Etiquetas separadas por coma")
	rootCmd.Flags().IntVarP(&severity, "severity", "s", 3, "Nivel de severidad (1 a 5)")

	rootCmd.AddCommand(
		newListCmd(svc),
		newRetroCmd(svc),
		newUICmd(svc),
		newShellInitCmd(),
		newNukeCmd(svc),
	)

	return rootCmd.Execute()
}

func newListCmd(svc *rant.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Lista y descifra todos los registros guardados",
		RunE: func(cmd *cobra.Command, args []string) error {
			rants, err := svc.List(cmd.Context())
			if err != nil {
				return err
			}

			if len(rants) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No hay registros todavía.")
				return nil
			}

			for _, item := range rants {
				tagLabel := ""
				if len(item.Tags) > 0 {
					tagLabel = " #" + strings.Join(item.Tags, " #")
				}
				fmt.Fprintf(
					cmd.OutOrStdout(),
					"[%s] (S:%d)%s\n%s\n\n",
					item.CreatedAt.Local().Format("2006-01-02 15:04"),
					item.Severity,
					tagLabel,
					item.Content,
				)
			}
			return nil
		},
	}
}

func newRetroCmd(svc *rant.Service) *cobra.Command {
	var days int

	cmd := &cobra.Command{
		Use:   "retro",
		Short: "Muestra métricas objetivas de fricción sin exponer texto crudo",
		RunE: func(cmd *cobra.Command, args []string) error {
			metrics, err := svc.GetMetrics(cmd.Context(), days)
			if err != nil {
				return err
			}

			if metrics.TotalCount == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No hay datos suficientes en el período seleccionado.")
				return nil
			}

			windowLabel := "histórico completo"
			if days > 0 {
				windowLabel = fmt.Sprintf("últimos %d días", days)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Resumen (%s)\n", windowLabel)
			fmt.Fprintf(cmd.OutOrStdout(), "  Eventos totales:    %d\n", metrics.TotalCount)
			fmt.Fprintf(cmd.OutOrStdout(), "  Severidad promedio: %.2f / 5.00\n", metrics.AverageSeverity)
			fmt.Fprintf(cmd.OutOrStdout(), "  Día más crítico:    %s\n", dayNames[metrics.PeakDayOfWeek])
			fmt.Fprintf(cmd.OutOrStdout(), "  Franja pico:        %02d:00 - %02d:59\n", metrics.PeakHourOfDay, metrics.PeakHourOfDay)

			if len(metrics.TopTags) > 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "\nTop triggers:")
				limit := len(metrics.TopTags)
				if limit > 5 {
					limit = 5
				}
				for _, t := range metrics.TopTags[:limit] {
					fmt.Fprintf(cmd.OutOrStdout(), "  #%-16s %d eventos (impacto: %d, prom: %.1f)\n", t.Tag, t.Count, t.TotalSeverity, t.AvgSeverity)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVarP(&days, "days", "d", 30, "Cantidad de días hacia atrás (0 = todo)")
	return cmd
}

func newUICmd(svc *rant.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "ui",
		Short: "Abre el dashboard web local en el navegador",
		RunE: func(cmd *cobra.Command, args []string) error {
			return web.Serve(cmd.Context(), svc)
		},
	}
}

func newShellInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "shell-init",
		Short: "Imprime la configuración para evitar guardar comandos en el historial",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "setopt HIST_IGNORE_SPACE 2>/dev/null || HISTCONTROL=ignorespace")
			fmt.Fprintln(cmd.OutOrStdout(), "alias rant=' rant'")
		},
	}
}

func newNukeCmd(svc *rant.Service) *cobra.Command {
	return &cobra.Command{
		Use:   "nuke",
		Short: "Sobreescribe la base de datos y borra la clave maestra del Keychain",
		RunE: func(cmd *cobra.Command, args []string) error {
			fmt.Fprint(cmd.OutOrStdout(), "Escribí NUKE para destruir toda la bóveda de forma irreversible: ")
			reader := bufio.NewReader(cmd.InOrStdin())
			confirmation, err := reader.ReadString('\n')
			if err != nil {
				return err
			}

			if strings.TrimSpace(confirmation) != "NUKE" {
				fmt.Fprintln(cmd.OutOrStdout(), "Operación cancelada.")
				return nil
			}

			if err := svc.Nuke(); err != nil {
				return err
			}

			fmt.Fprintln(cmd.OutOrStdout(), "Bóveda y clave maestra eliminadas.")
			return nil
		},
	}
}

func captureFromEditor() (string, error) {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		if runtime.GOOS == "windows" {
			editor = "notepad"
		} else {
			editor = "vi"
		}
	}

	tmpFile, err := os.CreateTemp("", "rant-*.txt")
	if err != nil {
		return "", err
	}
	tmpPath := tmpFile.Name()
	_ = tmpFile.Close()

	defer wipeTempFile(tmpPath)

	parts := strings.Fields(editor)
	args := append(parts[1:], tmpPath)

	editorCmd := exec.Command(parts[0], args...)
	editorCmd.Stdin = os.Stdin
	editorCmd.Stdout = os.Stdout
	editorCmd.Stderr = os.Stderr

	if err := editorCmd.Run(); err != nil {
		return "", fmt.Errorf("run editor: %w", err)
	}

	content, err := os.ReadFile(tmpPath)
	if err != nil {
		return "", err
	}

	return string(content), nil
}

func wipeTempFile(path string) {
	info, err := os.Stat(path)
	if err == nil && info.Size() > 0 {
		if file, openErr := os.OpenFile(path, os.O_WRONLY, 0o600); openErr == nil {
			zeros := make([]byte, info.Size())
			_, _ = file.WriteAt(zeros, 0)
			_ = file.Sync()
			_ = file.Close()
		}
	}
	_ = os.Remove(path)
}
