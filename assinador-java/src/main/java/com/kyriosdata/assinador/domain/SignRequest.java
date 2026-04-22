package com.kyriosdata.assinador.domain;

/**
 * Representa os dados da requisição para operação real de criação de assinatura digital.
 * 
 * <p>Esta classe atua como um DTO para transportar informações de criação de assinatura.</p>
 */
public class SignRequest {
    
    private String payload;
    private String keyType; // "TOKEN", "PEM", "PKCS12"
    private String keyPath;
    private String pinOrPassword;

    public SignRequest() {}

    public String getPayload() {
        return payload;
    }

    public void setPayload(String payload) {
        this.payload = payload;
    }

    public String getKeyType() {
        return keyType;
    }

    public void setKeyType(String keyType) {
        this.keyType = keyType;
    }

    public String getKeyPath() {
        return keyPath;
    }

    public void setKeyPath(String keyPath) {
        this.keyPath = keyPath;
    }

    public String getPinOrPassword() {
        return pinOrPassword;
    }

    public void setPinOrPassword(String pinOrPassword) {
        this.pinOrPassword = pinOrPassword;
    }
}
