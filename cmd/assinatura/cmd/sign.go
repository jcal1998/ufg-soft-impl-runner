package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/jcal1998/ufg-soft-impl-runner/internal/fhir"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/state"
	"github.com/jcal1998/ufg-soft-impl-runner/internal/ui"
	"github.com/spf13/cobra"
)

var bundlePath string
var pin string
var cpf string

var signCmd = &cobra.Command{
	Use:   "sign",
	Short: "Assina um pacote FHIR (Bundle) utilizando o motor criptográfico local",
	Run: func(cmd *cobra.Command, args []string) {
		if bundlePath == "" {
			ui.Error("Caminho do pacote (--bundle) é obrigatório.")
			return
		}
		if pin == "" {
			ui.Error("PIN do token (--pin) é obrigatório.")
			return
		}
		if cpf == "" {
			cpf = "00000000000" // Default for testing
		}

		st, err := state.Load()
		if err != nil || st.Assinador == nil || st.Assinador.PID == 0 {
			ui.Error("Motor criptográfico não está em execução. Execute 'assinatura start' primeiro.")
			return
		}

		ui.Info(fmt.Sprintf("Lendo pacote FHIR: %s", bundlePath))
		content, err := os.ReadFile(bundlePath)
		if err != nil {
			ui.Error(fmt.Sprintf("Falha ao ler arquivo: %v", err))
			return
		}

		// Extrair ID do bundle
		var bundle map[string]interface{}
		if err := json.Unmarshal(content, &bundle); err != nil {
			ui.Error("Arquivo não é um JSON válido.")
			return
		}

		bundleID, _ := bundle["id"].(string)
		if bundleID == "" {
			bundleID = "unknown-id"
		}

		ui.Info("Gerando recurso Provenance e calculando Hash SHA-256...")
		provenanceJSON, err := fhir.GenerateProvenance(content, bundleID, cpf)
		if err != nil {
			ui.Error(fmt.Sprintf("Falha ao gerar Provenance: %v", err))
			return
		}

		ui.Info("Enviando requisição de assinatura para o Motor Java...")
		
		reqPayload := map[string]string{
			"payload":       provenanceJSON,
			"keyType":       "TOKEN",
			"pinOrPassword": pin,
		}
		reqBytes, _ := json.Marshal(reqPayload)

		url := fmt.Sprintf("http://localhost:%d/sign", st.Assinador.Port)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
		if err != nil {
			ui.Error(fmt.Sprintf("Falha na comunicação com o Motor: %v", err))
			return
		}
		defer resp.Body.Close()

		respBody, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != 200 {
			ui.Error(fmt.Sprintf("O motor retornou erro HTTP %d: %s", resp.StatusCode, string(respBody)))
			return
		}

		var sigResp map[string]interface{}
		if err := json.Unmarshal(respBody, &sigResp); err != nil {
			ui.Error("Falha ao interpretar resposta do motor.")
			return
		}

		valid, _ := sigResp["valid"].(bool)
		if !valid {
			msg, _ := sigResp["message"].(string)
			ui.Error(fmt.Sprintf("Falha na assinatura: %s", msg))
			return
		}

		jwsStr, _ := sigResp["signature"].(string)
		if jwsStr == "" {
			ui.Error("O motor não devolveu um JWS válido.")
			return
		}

		ui.Info("Assinatura JWS recebida com sucesso. Injetando no arquivo original...")

		// Inject JWS into Bundle signature array
		signatureObj := map[string]interface{}{
			"type": []map[string]interface{}{
				{
					"system": "urn:iso-astm:E1762-95:2013",
					"code":   "1.2.840.10065.1.12.1.1",
				},
			},
			"when":      bundle["timestamp"], // Optionally use current time
			"sigFormat": "application/jose",
			"data":      jwsStr,
		}

		bundle["signature"] = []interface{}{signatureObj}

		finalBytes, _ := json.MarshalIndent(bundle, "", "  ")
		
		outPath := bundlePath[:len(bundlePath)-len(".json")] + "-assinado.json"
		err = os.WriteFile(outPath, finalBytes, 0644)
		if err != nil {
			ui.Error(fmt.Sprintf("Falha ao salvar arquivo assinado: %v", err))
			return
		}

		ui.Success(fmt.Sprintf("Pacote assinado e salvo em: %s", outPath))
	},
}

func init() {
	rootCmd.AddCommand(signCmd)
	signCmd.Flags().StringVarP(&bundlePath, "bundle", "b", "", "Caminho para o arquivo FHIR (.json)")
	signCmd.Flags().StringVarP(&pin, "pin", "p", "", "PIN do Token SoftHSM2")
	signCmd.Flags().StringVarP(&cpf, "cpf", "c", "", "CPF do assinante (opcional)")
	signCmd.MarkFlagRequired("bundle")
	signCmd.MarkFlagRequired("pin")
}
