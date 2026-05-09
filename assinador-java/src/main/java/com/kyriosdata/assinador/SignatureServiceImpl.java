package com.kyriosdata.assinador;

import com.kyriosdata.assinador.crypto.MaterialCriptografico;
import com.kyriosdata.assinador.crypto.MaterialCriptograficoSigner;
import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.JWSHeader;
import com.nimbusds.jose.JWSObject;
import com.nimbusds.jose.Payload;
import com.nimbusds.jose.util.Base64;
import java.security.cert.X509Certificate;
import java.util.Collections;

public class SignatureServiceImpl implements SignatureService {
    private static final Logger logger = LoggerFactory.getLogger(SignatureServiceImpl.class);

    public SignatureServiceImpl() {
    }

    @Override
    public SignatureResponse sign(SignRequest request) {
        if (request.getPayload() == null || request.getPayload().isEmpty()) {
            return new SignatureResponse(null, false, "O Payload não pode ser vazio.");
        }

        try {
            MaterialCriptografico mc = null;
            
            if ("TOKEN".equalsIgnoreCase(request.getKeyType())) {
                // Instancia o MC com os dados básicos
                mc = new MaterialCriptografico(
                    request.getPinOrPassword() != null ? request.getPinOrPassword().toCharArray() : new char[0],
                    "SoftHSM",
                    0,
                    "Token1"
                );
            } else {
                return new SignatureResponse(null, false, "O tipo de chave " + request.getKeyType() + " ainda não está implementado.");
            }

            if (mc == null || !mc.isAvailable()) {
                return new SignatureResponse(null, false, "Material Criptográfico não pôde ser carregado ou não está disponível.");
            }

            X509Certificate certificate = mc.getPublicKey();

            // Construir o Header do JWS conforme padrão (RS256 + x5c com o certificado)
            JWSHeader header = new JWSHeader.Builder(JWSAlgorithm.RS256)
                    .x509CertChain(Collections.singletonList(Base64.encode(certificate.getEncoded())))
                    .build();

            // O Payload é o JSON do Provenance (montado pelo cliente/CLI)
            Payload payload = new Payload(request.getPayload());

            // Juntar no objeto JWS e assinar usando o custom JWSSigner que invoca o MC
            JWSObject jwsObject = new JWSObject(header, payload);
            MaterialCriptograficoSigner signer = new MaterialCriptograficoSigner(mc);
            jwsObject.sign(signer);
            
            // Serializar no formato Compacto (Header.Payload.Signature)
            String jwsString = jwsObject.serialize();
            
            logger.info("Assinatura JWS gerada com sucesso via Material Criptográfico (SoftHSM2).");
            return new SignatureResponse(jwsString, true, "JWS Assinado com sucesso pelo dispositivo.");

        } catch (Exception e) {
            logger.error("Falha ao assinar payload: ", e);
            return new SignatureResponse(null, false, "Falha interna no motor criptográfico: " + e.getMessage());
        }
    }

    @Override
    public SignatureResponse validate(ValidateRequest request) {
        return new SignatureResponse(null, false, "A validação (Validate) ainda não está implementada no Motor.");
    }
}
