package com.kyriosdata.assinador;

import com.kyriosdata.assinador.crypto.Pkcs11TokenService;
import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.security.KeyStore;
import java.security.PrivateKey;
import java.security.Signature;
import java.security.cert.Certificate;
import java.util.Base64;
import java.util.Enumeration;

public class SignatureServiceImpl implements SignatureService {
    private static final Logger logger = LoggerFactory.getLogger(SignatureServiceImpl.class);
    private final Pkcs11TokenService tokenService;

    public SignatureServiceImpl() {
        this.tokenService = new Pkcs11TokenService();
    }

    @Override
    public SignatureResponse sign(SignRequest request) {
        if (request.getPayload() == null || request.getPayload().isEmpty()) {
            return new SignatureResponse(null, false, "O Payload não pode ser vazio.");
        }

        try {
            KeyStore keyStore;
            PrivateKey privateKey = null;
            
            if ("TOKEN".equalsIgnoreCase(request.getKeyType())) {
                keyStore = tokenService.loadHardwareToken(request.getPinOrPassword(), null);
                
                // Pega a primeira chave privada encontrada no token
                Enumeration<String> aliases = keyStore.aliases();
                while (aliases.hasMoreElements()) {
                    String alias = aliases.nextElement();
                    if (keyStore.isKeyEntry(alias)) {
                        privateKey = (PrivateKey) keyStore.getKey(alias, null); // PKCS11 usually ignores the password here
                        logger.info("Chave privada encontrada no Token (Alias: {})", alias);
                        break;
                    }
                }
            } else {
                // Futura implementação para PEM ou PKCS12
                return new SignatureResponse(null, false, "O tipo de chave " + request.getKeyType() + " ainda não está implementado no Backend.");
            }

            if (privateKey == null) {
                return new SignatureResponse(null, false, "Nenhuma chave privada foi encontrada no dispositivo de assinatura.");
            }

            // Gerar a assinatura bruta (Raw Signature) usando SHA256withRSA
            Signature signature = Signature.getInstance("SHA256withRSA");
            signature.initSign(privateKey);
            signature.update(request.getPayload().getBytes("UTF-8"));
            
            byte[] digitalSignature = signature.sign();
            String base64Signature = Base64.getEncoder().encodeToString(digitalSignature);
            
            logger.info("Assinatura gerada com sucesso.");
            return new SignatureResponse(base64Signature, true, "Assinado com sucesso pelo dispositivo de hardware.");

        } catch (Exception e) {
            logger.error("Falha ao assinar payload: ", e);
            return new SignatureResponse(null, false, "Falha interna no motor criptográfico: " + e.getMessage());
        }
    }

    @Override
    public SignatureResponse validate(ValidateRequest request) {
        // Futuro: Validar assinatura com a chave pública
        return new SignatureResponse(null, false, "A validação (Validate) ainda não está implementada no Motor.");
    }
}
