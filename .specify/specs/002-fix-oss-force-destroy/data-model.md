# Data Model: OSS Bucket

## Entities

### OSS Bucket
- **Id**: String (Bucket Name)
- **force_destroy**: Boolean (Trigger for recursive deletion)
- **Tags**: Map (Resource tags)

### OSS Object (Implicit)
- **Key**: String
- **VersionId**: String
- **IsDeleteMarker**: Boolean

### OSS Multipart Upload (Implicit)
- **UploadId**: String
- **Key**: String
