package platformauthority

import gcpkmsadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/gcpkms"

type KMSClient = gcpkmsadapter.Client
type KMSPublicKey = gcpkmsadapter.PublicKey
type KMSAsymmetricSignRequest = gcpkmsadapter.AsymmetricSignRequest
type KMSAsymmetricSignResponse = gcpkmsadapter.AsymmetricSignResponse

const KMSRequiredIAMPermission = gcpkmsadapter.RequiredIAMPermission
