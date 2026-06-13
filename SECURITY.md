# Security Policy

## Reporting a Vulnerability

If you discover any security-related issues or vulnerabilities, please contact us at [n.chika156@gmail.com](mailto:n.chika156@gmail.com). We appreciate your responsible disclosure and will work with you to address the issue promptly.

## Handling of Personal Data

`career` processes personal information (name, address, phone number, work
history) found in your resume YAML file:

- All input is read from a local YAML file and written to a local PDF file. The
  tool makes no network connections and sends nothing anywhere.
- Your resume YAML contains sensitive personal data. Keep it out of public
  repositories, and prefer a private location or a local-only file.
- Generated PDFs are equally sensitive. The bundled `.gitignore` ignores `*.pdf`
  so they are not committed by accident.

## Supported Versions

We recommend using the latest release for the most up-to-date and secure experience. Security updates are provided for the latest stable version.

## Security Policy

- Security issues are treated with the highest priority.
- We follow responsible disclosure practices.
- Fixes for security vulnerabilities will be provided in a timely manner.

## Acknowledgments

We would like to thank the security researchers and contributors who responsibly report security issues and work with us to make our project more secure.

Thank you for your help in making our project safe and secure for everyone.
