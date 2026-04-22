package com.kyriosdata.assinador.server;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.kyriosdata.assinador.SignatureService;
import com.kyriosdata.assinador.domain.SignRequest;
import com.kyriosdata.assinador.domain.SignatureResponse;
import com.kyriosdata.assinador.domain.ValidateRequest;
import io.javalin.Javalin;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class App {
    private static final Logger logger = LoggerFactory.getLogger(App.class);
    private static final ObjectMapper mapper = new ObjectMapper();
    
    // Use the real Signature Service with SunPKCS11 integration
    private static SignatureService signatureService = new com.kyriosdata.assinador.SignatureServiceImpl();

    public static void main(String[] args) {
        int port = 8081; // Default port
        
        // Parse port from args if provided: --server.port=8080
        for (String arg : args) {
            if (arg.startsWith("--server.port=")) {
                try {
                    port = Integer.parseInt(arg.substring("--server.port=".length()));
                } catch (NumberFormatException e) {
                    logger.warn("Invalid port provided in arguments. Falling back to default: {}", port);
                }
            }
        }

        Javalin app = Javalin.create(config -> {
            config.showJavalinBanner = false;
        }).start(port);

        logger.info("Assinador Backend started on port {}", port);

        // Health check endpoint
        app.get("/health", ctx -> {
            ctx.status(200).result("OK");
        });

        // Signing endpoint
        app.post("/sign", ctx -> {
            try {
                SignRequest request = mapper.readValue(ctx.body(), SignRequest.class);
                SignatureResponse response = signatureService.sign(request);
                ctx.status(200).json(response);
            } catch (Exception e) {
                logger.error("Error signing request", e);
                ctx.status(400).json(new SignatureResponse(null, false, "Invalid Input: " + e.getMessage()));
            }
        });

        // Validation endpoint
        app.post("/validate", ctx -> {
            try {
                ValidateRequest request = mapper.readValue(ctx.body(), ValidateRequest.class);
                SignatureResponse response = signatureService.validate(request);
                ctx.status(200).json(response);
            } catch (Exception e) {
                logger.error("Error validating request", e);
                ctx.status(400).json(new SignatureResponse(null, false, "Invalid Input: " + e.getMessage()));
            }
        });
    }
}
