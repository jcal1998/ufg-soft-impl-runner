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
import java.util.Base64;

public class SignatureServiceImpl implements SignatureService {
    private static final Logger logger = LoggerFactory.getLogger(SignatureServiceImpl.class);
    private final Pkcs11TokenService tokenService = new Pkcs11TokenService();

    @Override
    public SignatureResponse sign(SignRequest request) {
        try {
            logger.info("Iniciando assinatura. Tipo: {}", request.getKeyType());

            KeyStore keyStore = null;

            if ("TOKEN".equalsIgnoreCase(request.getKeyType())) {
                keyStore = tokenService.loadHardwareToken(request.getPinOrPassword(), request.getKeyPath());
            } else {
                return new SignatureResponse(null, false, "Tipo de chave " + request.getKeyType() + " ainda não suportado no backend.");
            }

            // Find the first alias (Hardware tokens usually have 1 or 2 aliases)
            String alias = null;
            java.util.Enumeration<String> aliases = keyStore.aliases();
            while (aliases.hasMoreElements()) {
                String a = aliases.nextElement();
                if (keyStore.isKeyEntry(a)) {
                    alias = a;
                    break;
                }
            }

            if (alias == null) {
                return new SignatureResponse(null, false, "Nenhuma chave privada encontrada no token.");
            }

            PrivateKey privateKey = (PrivateKey) keyStore.getKey(alias, null);

            // Execute the RSA-SHA256 Signature
            Signature signatureAlgorithm = Signature.getInstance("SHA256withRSA");
            signatureAlgorithm.initSign(privateKey);
            signatureAlgorithm.update(request.getPayload().getBytes("UTF-8"));

            byte[] signatureBytes = signatureAlgorithm.sign();
            String base64Signature = Base64.getEncoder().encodeToString(signatureBytes);

            logger.info("Assinatura gerada com sucesso.");
            return new SignatureResponse(base64Signature, true, "Assinatura gerada com SunPKCS11");

        } catch (Exception e) {
            logger.error("Erro fatal ao assinar", e);
            return new SignatureResponse(null, false, "Erro ao assinar: " + e.getMessage());
        }
    }

    @Override
    public SignatureResponse validate(ValidateRequest request) {
        return new SignatureResponse(null, false, "Validação ainda não implementada.");
    }
}
