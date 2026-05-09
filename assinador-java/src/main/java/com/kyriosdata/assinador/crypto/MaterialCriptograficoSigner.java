package com.kyriosdata.assinador.crypto;

import com.nimbusds.jose.JOSEException;
import com.nimbusds.jose.JWSAlgorithm;
import com.nimbusds.jose.JWSHeader;
import com.nimbusds.jose.JWSSigner;
import com.nimbusds.jose.jca.JCAContext;
import com.nimbusds.jose.util.Base64URL;

import java.util.Collections;
import java.util.Set;

public class MaterialCriptograficoSigner implements JWSSigner {

    private final MaterialCriptografico mc;
    private final JCAContext jcaContext = new JCAContext();

    public MaterialCriptograficoSigner(MaterialCriptografico mc) {
        this.mc = mc;
    }

    @Override
    public Base64URL sign(JWSHeader header, byte[] signingInput) throws JOSEException {
        try {
            byte[] signature = mc.sign(signingInput);
            return Base64URL.encode(signature);
        } catch (Exception e) {
            throw new JOSEException("Falha ao assinar payload via Material Criptografico", e);
        }
    }

    @Override
    public Set<JWSAlgorithm> supportedJWSAlgorithms() {
        return Collections.singleton(JWSAlgorithm.RS256);
    }

    @Override
    public JCAContext getJCAContext() {
        return jcaContext;
    }
}
