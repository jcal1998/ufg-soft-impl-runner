package com.kyriosdata.assinador.crypto;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.io.File;
import java.security.KeyStore;
import java.security.Provider;
import java.security.Security;
import java.util.Arrays;
import java.util.List;

public class Pkcs11TokenService {
    private static final Logger logger = LoggerFactory.getLogger(Pkcs11TokenService.class);

    // List of known driver paths for SafeNet and Pronova across different OS
    private static final List<String> KNOWN_DRIVERS = Arrays.asList(
            // Windows
            "C:\\Windows\\System32\\eTPkcs11.dll",
            "C:\\Windows\\System32\\eps2003csp11.dll",
            // Mac
            "/usr/local/lib/libeToken.dylib",
            "/Library/Frameworks/eToken.framework/Versions/A/libeToken.dylib",
            // Linux
            "/usr/lib/libeToken.so",
            "/usr/lib/libeps2003.so"
    );

    /**
     * Detects the first available PKCS#11 driver in the system
     */
    public String findAvailableDriver() {
        for (String driverPath : KNOWN_DRIVERS) {
            File f = new File(driverPath);
            if (f.exists() && f.isFile()) {
                logger.info("PKCS#11 Driver found: {}", driverPath);
                return driverPath;
            }
        }
        logger.warn("No compatible PKCS#11 driver found on this system.");
        return null;
    }

    /**
     * Loads the Hardware Token KeyStore using the provided PIN
     */
    public KeyStore loadHardwareToken(String pin, String explicitDriverPath) throws Exception {
        String driverPath = explicitDriverPath;
        if (driverPath == null || driverPath.isEmpty()) {
            driverPath = findAvailableDriver();
        }

        if (driverPath == null) {
            throw new IllegalStateException("Nenhum driver de Token USB/Smartcard foi encontrado. Verifique se o SafeNet ou Pronova estão instalados.");
        }

        String config = "--name=HardwareToken\nlibrary=" + driverPath;
        
        // Java 9+ way of initializing SunPKCS11
        Provider pkcs11Provider = Security.getProvider("SunPKCS11");
        if (pkcs11Provider == null) {
            throw new IllegalStateException("Provedor SunPKCS11 não disponível nesta JVM.");
        }
        
        pkcs11Provider = pkcs11Provider.configure(config);
        Security.addProvider(pkcs11Provider);

        KeyStore keyStore = KeyStore.getInstance("PKCS11", pkcs11Provider);
        keyStore.load(null, pin != null ? pin.toCharArray() : new char[0]);
        
        logger.info("Hardware Token KeyStore loaded successfully.");
        return keyStore;
    }
}
