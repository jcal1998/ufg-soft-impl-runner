package fhir

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"
)

// Provenance representa o recurso FHIR Provenance que será assinado
type Provenance struct {
	ResourceType string      `json:"resourceType"`
	Target       []Reference `json:"target"`
	Recorded     string      `json:"recorded"`
	Agent        []Agent     `json:"agent"`
}

type Reference struct {
	Reference string `json:"reference"`
}

type Agent struct {
	Type Type `json:"type"`
	Who  Who  `json:"who"`
}

type Type struct {
	Coding []Coding `json:"coding"`
}

type Coding struct {
	System string `json:"system"`
	Code   string `json:"code"`
}

type Who struct {
	Identifier Identifier `json:"identifier"`
}

type Identifier struct {
	System string `json:"system"`
	Value  string `json:"value"`
}

// GenerateProvenance cria o objeto JSON do Provenance baseado no hash do bundle
func GenerateProvenance(bundleContent []byte, bundleID string, cpf string) (string, error) {
	// Calcular Hash SHA-256 do Bundle
	hash := sha256.Sum256(bundleContent)
	hashHex := hex.EncodeToString(hash[:])

	// Na especificação FHIR/SES-GO, o target reference pode apontar para o UUID do bundle e incluir o hash de alguma forma
	// Por simplicidade, assumimos que o target referência é urn:uuid:ID
	targetRef := fmt.Sprintf("urn:uuid:%s", bundleID)
	
	// A documentação pede que o conteúdo a ser assinado seja o hash da canonicalização.
	// Vamos colocar o hash como um target extension ou diretamente como parte do Provenance para ser assinado no Payload JWS.
	// Aqui usaremos o Target reference e adicionaremos um texto indicativo do Hash.

	now := time.Now().UTC().Format(time.RFC3339)

	prov := Provenance{
		ResourceType: "Provenance",
		Target: []Reference{
			{Reference: targetRef},
			{Reference: fmt.Sprintf("urn:hash:sha256:%s", hashHex)}, // Injetando o hash como referência adicional
		},
		Recorded: now,
		Agent: []Agent{
			{
				Type: Type{
					Coding: []Coding{
						{
							System: "http://terminology.hl7.org/CodeSystem/provenance-participant-type",
							Code:   "author",
						},
					},
				},
				Who: Who{
					Identifier: Identifier{
						System: "http://receita.fazenda.gov.br",
						Value:  cpf,
					},
				},
			},
		},
	}

	jsonBytes, err := json.Marshal(prov)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
