# Package distribution guide
## MacOS

1. Create a new Developer ID Application certificate in https://developer.apple.com/account/resources/certificates/list
2. To be able to obtain the certificate, you need to create a Certificate Signing Request (CSR). Open `Keychain Access` and go to `Certificate Assistant` -> `Request a Certificate from a Certificate Authority`. Select your request to be `Saved to disk`. Specify email address of the developer account.
3. Upload the CSR request file (`CertificateSigningRequest.certSigningRequest`). Download the certificate (`developerID_application.cer`), import it to the `Keychain Access`.
4. Go to `My Certificates` in `Keychain Access`, select both the downloaded certificate and the private key and select `Export 2 items...`, choose `Personal Information Exchange (.p12)` format. Save the file as `certificates.p12`. Store the password used for exporting in `MACOS_CERTIFICATE_PASSWORD` env variable.
5. Get base64 version of the certificate by entering in the commandline `base64 certificates.p12 | pbcopy`. Store the result in `MACOS_CERTIFICATE` env variable.
6. Generate application password in https://appleid.apple.com/account/manage and store it as `MACOS_APPLICATION_PASSWORD`.
7. Store Apple ID username (developer account email address) in `MACOS_APPLE_ID_USERNAME`
8. Store the ID of the Developer ID Application certificate in `MACOS_CERTIFICATE_IDENTITY`, format: `"Developer ID Application: {name} ({id})"`.
