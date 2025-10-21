# Feature Specification: Logged-In Users Report Damaged Roads

**Feature Branch**: `002-logged-in-user`  
**Created**: October 19, 2025  
**Status**: Draft  
**Input**: User description: "logged in user should be able to report damaged road easily. the report can consist title, subdistrict code of administrative area code in indonesia, array of latitude longitude pair describing damaged report path, array of photos url, description (optional)."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Submit a new damaged road report (Priority: P1)

A logged-in resident captures the essential details of a damaged road and submits the report so local authorities receive actionable information.

**Why this priority**: Collecting the first report is the core value of the feature and unlocks the data pipeline for downstream teams.

**Independent Test**: Verify that an authenticated user can submit a report with the required fields and receive confirmation that the report was recorded.

**Acceptance Scenarios**:

1. **Given** the user is authenticated and has required details, **When** they provide a title, subdistrict code, at least one valid coordinate, at least one photo URL, and submit, **Then** the system records the report and confirms receipt.
2. **Given** the user omits any mandatory field such as the coordinate or photo, **When** they attempt submission, **Then** the system blocks the submission and explains which field must be provided.

---

### User Story 2 - Describe the damage path accurately (Priority: P2)

A logged-in resident traces the road segment that is damaged so responders can locate the issue precisely.

**Why this priority**: Accurate location data shortens inspection time and prevents misdirected responses.

**Independent Test**: Verify that a user can supply at least one coordinate pair (with optional additional points) forming the damaged path and that the path is stored with the report.

**Acceptance Scenarios**:

1. **Given** the user provides one or more coordinate pairs within the targeted subdistrict, **When** they submit the report, **Then** the stored report includes the ordered path so responders can map the damage.
2. **Given** the user provides coordinates that are clearly outside Indonesia, **When** they attempt submission, **Then** the system declines the request and explains the coordinates must fall within national boundaries.

---

### User Story 3 - Add supporting evidence (Priority: P3)

A logged-in resident attaches photos and contextual description to improve report credibility and prioritization.

**Why this priority**: Supporting evidence helps authorities triage and validate reports without extra fieldwork.

**Independent Test**: Verify that required photo evidence and optional text can be added, stored with the report, and retrieved for later review.

**Acceptance Scenarios**:

1. **Given** the user adds at least one valid photo URL (with optional additional photos) and an optional description, **When** they submit the report, **Then** the system stores the attachments alongside the report metadata.
2. **Given** the user provides a photo URL that is not accessible, **When** they attempt submission, **Then** the system flags the invalid link and prompts the user to correct it before continuing.

### Edge Cases

- Missing coordinates: no coordinate pairs supplied results in a blocked submission with guidance to add at least one point.
- Missing photo evidence: no photo URLs supplied results in a blocked submission with instructions to capture or upload at least one image.
- Photo limit exceeded: attempts to add more than 10 photo URLs results in a blocked submission with explanation of the maximum limit.
- Out-of-range geospatial data: coordinates outside Indonesian boundaries trigger a validation error and message.
- Unsupported or unreachable photo URLs: submission is paused until the user removes or updates the problematic links.
- Duplicate submissions: if a user attempts to submit a report with identical title, location, and timestamp within a short window, the system warns them and encourages checking existing reports.
- Session timeout mid-entry: if the session expires before submission, the user is prompted to reauthenticate without losing entered data.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Only authenticated users may access the damaged road reporting capability.
- **FR-002**: The system must collect and require a title summarizing the damage, a valid Kemendagri hierarchical subdistrict (desa/kelurahan) code (e.g., `35.10.02.2005`), and an ordered list containing at least one coordinate pair describing the affected location or path.
- **FR-003**: The system must allow an optional free-text description up to 500 characters to provide additional context.
- **FR-004**: The system must require at least one supporting photo URL and allow a maximum of 10 photo URLs per report, validating that each URL is well-formed and accessible before submission completes. **Security Requirements**: (a) Only HTTP and HTTPS protocols are allowed; (b) URLs must not point to localhost, private IP ranges (10.x.x.x, 172.16-31.x.x, 192.168.x.x), or link-local addresses to prevent SSRF attacks; (c) URL accessibility checks must timeout after 5 seconds; (d) Only image content types (image/jpeg, image/png, image/webp) are accepted.
- **FR-005**: One or more coordinate points describing the damaged path must be provided, with each latitude-longitude pair falling within Indonesian national boundaries (latitude: -11 to 6, longitude: 95 to 141). The order of points is preserved to reflect the damaged road segment.
- **FR-006**: The system must validate that the SubDistrictCode (Kemendagri format: NN.NN.NN.NNNN) exists in the official administrative dataset and that at least one coordinate from the path falls within 200 meters of the subdistrict's geographic centroid. **Implementation Note**: Centroid coordinates should be sourced from official Indonesian government geospatial data (BIG - Badan Informasi Geospasial) or cached administrative boundary polygons with calculated centroids. Validation uses Haversine distance formula for proximity checking.
- **FR-007**: Upon successful submission, the system must persist the report with timestamps, author identifier, and all provided details for later review.
- **FR-008**: If validation fails, the system must explain the blocking issue in plain language without discarding previously entered information.
- **FR-009**: Users must be able to review a submission summary before final confirmation, including all entered fields and attachments.
- **FR-010**: The system must return a clear confirmation referencing the report identifier once the submission is stored.

### Key Entities *(include if feature involves data)*

- **Damaged Road**: Represents a single reported occurrence of road damage from a resident; includes title, Kemendagri administrative code, coordinate path, optional description, required photo evidence, author identifier, and timestamps.
-   - Attributes: `title`, `kemendagri_code` (province.district.subdistrict.village), `path_points[]`, `description?`, `photos[]` (>=1, <=10), `author_id`, `created_at`, `updated_at`.
- **Damaged Road Path**: Captures the ordered list of latitude-longitude pairs describing the damaged stretch; ensures minimum length (>=1 point) and geographic validity within administrative boundaries.
- **Damaged Road Photo Attachment**: Holds metadata for each required photo URL, including source link, optional caption, and validation status (reachable and properly formatted).

## Assumptions

- Users provide the Kemendagri hierarchical administrative code (province.district.subdistrict.village) for the damaged location; the system offers lookup assistance and formatting validation (pattern `NN.NN.NN.NNNN`). Example: `35.10.02.2005`.
- At least one coordinate point is sufficient to anchor the road damage location for initial response purposes, with additional points improving accuracy when available.
- At least one clear photo is required per report with a maximum of 10 photos allowed; users are alerted early so they can capture imagery before starting submission.
- Connectivity may be intermittent, so partially completed forms are preserved locally during validation.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 90% of authenticated users can submit a complete damaged road report (title, location, path) within 3 minutes on their first attempt.
- **SC-002**: 95% of stored reports include coordinate paths where at least one point falls within 200 meters of the selected subdistrict's geographic centroid (validated using Haversine distance calculation), indicating location accuracy.
- **SC-003**: At least 90% of submitted reports include required photo evidence that reviewers rate as sufficiently clear for triage during pilot evaluations.
- **SC-004**: 100% of reports enforce the 10-photo limit, with users receiving clear feedback when attempting to exceed this constraint.
- **SC-005**: Duplicate submission warnings reduce repeated reports for the same location within a 24-hour period by at least 50% compared to baseline.

## Clarifications

### Session 2025-10-19

- Q: What lifecycle/state model should a Damaged Road follow after submission? → A: Submitted → Under Verification → Verified → Resolved → Archived (Needs Info removed per clarification)

### Lifecycle Model

The Damaged Road entity progresses through these states:
1. **Submitted**: Initial creation by user; pending review.
2. **Under Verification**: Moderation or authority evaluating validity.
3. **Verified**: Sufficient evidence confirms the damage; eligible for resolution tracking.
4. **Resolved**: Authority indicates repair completed or damage no longer present.
5. **Archived**: Historical record retained for analytics; no further edits allowed.

Constraints:
- Transitions are strictly forward (no backward transitions).
- Verified requires evidence review completion.
- Resolved requires at least one verification action recorded plus resolution timestamp.
