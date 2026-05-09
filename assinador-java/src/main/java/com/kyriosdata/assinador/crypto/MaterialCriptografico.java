package com.kyriosdata.assinador.crypto;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.io.FileWriter;
import java.security.KeyStore;
import java.security.PrivateKey;
import java.security.Provider;
import java.security.Security;
import java.security.Signature;
import java.security.cert.Certificate;
import java.security.cert.X509Certificate;
import java.util.Enumeration;

public class MaterialCriptografico {
    private static final Logger logger = LoggerFactory.getLogger(MaterialCriptografico.class);

    private PrivateKey privateKey;
    private X509Certificate publicCertificate;
    private boolean available = false;

    public MaterialCriptografico(char[] pin, String id, int slotId, String tokenLabel) {
        try {
            // Find driver path
            String driverPath = getDriverPath();
            if (driverPath == null) {
                logger.error("Driver do SoftHSM2 não encontrado na máquina.");
                return;
            }

            // Onde está o .so ou .dll? Monta arquivo (MKTEMP)
            File tempConfig = File.createTempFile("pkcs11-", ".cfg");
            tempConfig.deleteOnExit();
            try (FileWriter fw = new FileWriter(tempConfig)) {
                fw.write("name = SoftHSM\n");
                fw.write("library = " + driverPath + "\n");
                //fw.write("slotListIndex = " + slotId + "\n"); // Optional configuration based on token mapping
            }

            // Load Provider via SunPKCS11
            Provider pkcs11Provider = Security.getProvider("SunPKCS11");
            if (pkcs11Provider == null) {
                logger.error("Provedor SunPKCS11 não disponível nesta JVM.");
                return;
            }
            pkcs11Provider = pkcs11Provider.configure(tempConfig.getAbsolutePath());
            Security.addProvider(pkcs11Provider);

            // Load KeyStore
            KeyStore keyStore = KeyStore.getInstance("PKCS11", pkcs11Provider);
            keyStore.load(null, pin);

            // Fetch Key and Certificate
            Enumeration<String> aliases = keyStore.aliases();
            while (aliases.hasMoreElements()) {
                String alias = aliases.nextElement();
                if (keyStore.isKeyEntry(alias)) {
                    this.privateKey = (PrivateKey) keyStore.getKey(alias, pin);
                    Certificate cert = keyStore.getCertificate(alias);
                    if (cert instanceof X509Certificate) {
                        this.publicCertificate = (X509Certificate) cert;
                        this.available = true;
                        logger.info("Chave privada carregada via SoftHSM2 (Alias: {})", alias);
                        break;
                    }
                }
            }

        } catch (Exception e) {
            logger.error("Erro ao inicializar Material Criptografico: ", e);
        }
    }

    private String getDriverPath() {
        String[] paths = {
                "/opt/homebrew/lib/softhsm/libsofthsm2.so", // Mac ARM Homebrew
                "/usr/local/lib/softhsm/libsofthsm2.so", // Mac Intel Homebrew
                "/usr/lib/softhsm/libsofthsm2.so", // Linux Debian/Ubuntu
                "/usr/lib/x86_64-linux-gnu/softhsm/libsofthsm2.so",
                "C:\\SoftHSM2\\lib\\softhsm2-x64.dll", // Windows x64
                "C:\\SoftHSM2\\lib\\softhsm2.dll" // Windows x86
        };
        for (String path : paths) {
            if (new File(path).exists()) {
                logger.info("Driver SoftHSM2 encontrado: {}", path);
                return path;
            }
        }
        return null;
    }

    public boolean isAvailable() {
        return this.available;
    }

    public byte[] sign(byte[] payload) throws Exception {
        if (!available) {
            throw new IllegalStateException("Material Criptografico indisponivel.");
        }
        Signature signature = Signature.getInstance("SHA256withRSA");
        signature.initSign(privateKey);
        signature.update(payload);
        return signature.sign();
    }

    public X509Certificate getPublicKey() {
        return publicCertificate;
    }
}
