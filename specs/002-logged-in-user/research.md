# Research Findings: Logged-In Users Report Damaged Roads

**Date**: 2025-10-19
**Branch**: 002-logged-in-user

## Performance Goals Research

### Decision: Target performance metrics for Indonesian road damage reporting system

**Rationale**: Based on Indonesian network conditions, mobile-first usage patterns, and regional deployment considerations.

**Performance Targets**:
- **Report Submission**: < 2 seconds (95th percentile), < 3 seconds (99th percentile)
- **Photo URL Validation**: < 500ms per URL, < 2 seconds for batch (10 photos)
- **Geospatial Validation**: < 300ms for boundary checks, < 800ms for admin code validation
- **Administrative Code Validation**: < 200ms format check, < 400ms existence verification

**Concurrency Targets**:
- **Pilot Phase**: 50-100 concurrent users, 20-50 requests/second
- **Regional Rollout**: 500-1,000 concurrent users, 200-400 requests/second
- **Nationwide Deployment**: 5,000-10,000 concurrent users, 2,000-4,000 requests/second

**Infrastructure Considerations**:
- Indonesian network latencies: 50-150ms (urban), 200-400ms (rural)
- Mobile device dominance: 85%+ Android usage
- Peak usage: 7-9 AM and 5-9 PM, weekends

**Alternatives Considered**:
- Stricter performance targets (< 1 second) - rejected due to Indonesian network realities
- Single deployment approach - rejected in favor of phased scaling

## Indonesian Geospatial Data Research

### Decision: Use official Indonesian government data sources with PostGIS validation

**Rationale**: Official sources provide accurate, up-to-date administrative boundaries with proper licensing for civic applications.

**Data Sources**:
- **Portal Satu Data Indonesia** (data.go.id) - Primary official data portal
- **TanahAir Indonesia** (tanahair.indonesia.go.id) - National geospatial portal by BIG
- **Kemendagri** - Ministry of Home Affairs administrative code registry

**Technical Implementation**:
- **Data Format**: Shapefile/GeoJSON stored in PostGIS with spatial indexing
- **Validation Approach**: Point-in-polygon queries with proximity fallbacks
- **Update Strategy**: Quarterly monitoring of official sources
- **Performance**: < 10ms local queries vs > 200ms online WFS calls

**Code Structure**: NN.NN.NN.NNNN (Province.District.Subdistrict.Village)
- Covers 34 provinces, ~500 districts, ~7,000 subdistricts, ~83,000 villages
- Regular changes due to pemekaran (regional formation)

**Alternatives Considered**:
- Online WFS services - rejected due to network dependency and performance
- Third-party commercial data - rejected due to licensing costs and update frequency

## Photo URL Validation Strategy Research

### Decision: Multi-layered validation with SSRF protection and Indonesian network optimization

**Rationale**: Balance security requirements with Indonesian network reliability and mobile usage patterns.

**Validation Approach**:
1. **HTTP HEAD requests** for basic accessibility check
2. **Content-Type validation** for image formats (JPEG, PNG, WebP, HEIC)
3. **Selective GET requests** for dimension verification
4. **SSRF protection** with private IP blocking and domain whitelisting
5. **Concurrent validation** with configurable limits

**Security Measures**:
- Private IP range blocking (127.0.0.0/8, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16)
- Allowed domains whitelist (Instagram, Facebook, WhatsApp, Indonesian platforms)
- Rate limiting per user (5 requests/minute with token bucket)
- File size limits (10MB for mobile photos)
- Suspicious pattern detection

**Performance Optimizations**:
- **Timeouts**: 30-second base timeout for Indonesian networks
- **Caching**: 1-hour TTL for validation results
- **Concurrency**: Configurable limit (default 5 concurrent validations)
- **Retry Logic**: Progressive timeouts with exponential backoff

**Indonesian Context**:
- **Allowed Domains**: Major Indonesian platforms (Tokopedia, Bukalapak, Shopee)
- **Network Optimization**: Longer timeouts, retry mechanisms for connectivity issues
- **Mobile Considerations**: Support for HEIC format (common on iOS)

**Alternatives Considered**:
- Full image download for validation - rejected due to bandwidth and privacy concerns
- Basic regex validation only - rejected due to security risks
- No validation (user trust) - rejected due to abuse potential

## Scale and Storage Requirements

### Decision: Phased scaling approach with PostgreSQL storage optimization

**Rationale**: Indonesian market size requires scalable but cost-effective deployment strategy.

**Storage Estimates**:
- **Single Report**: ~1KB (metadata) + 1-10 photo URLs (100B each) = ~2KB
- **Pilot Phase** (1,000 reports/month): ~2MB/month, ~24MB/year
- **Regional Rollout** (10,000 reports/month): ~20MB/month, ~240MB/year
- **Nationwide** (100,000 reports/month): ~200MB/month, ~2.4GB/year

**Database Strategy**:
- **PostgreSQL** with PostGIS for geospatial data
- **Table Partitioning** by date for large-scale deployments
- **Spatial Indexing** for coordinate queries
- **Connection Pooling** for high concurrency

**Infrastructure Scaling**:
- **Pilot**: 2 CPU, 4GB RAM, 50GB storage
- **Regional**: 4 CPU, 8GB RAM, 200GB storage
- **Nationwide**: 8+ CPU, 16GB+ RAM, 1TB+ storage

**Alternatives Considered**:
- NoSQL database - rejected due to geospatial query requirements
- Cloud-native storage - rejected due to data sovereignty requirements
- Single-tier scaling - rejected due to cost considerations

## Technology Stack Confirmation

### Decision: Standardize on constitution-mandated technology stack

**Confirmed Technologies**:
- **Go 1.21+** - Type safety, performance, ecosystem
- **Gin Framework** - HTTP routing and middleware
- **PostgreSQL** - Primary database with PostGIS
- **JWT** - Stateless authentication
- **Swagger/OpenAPI** - API documentation

**Additional Libraries**:
- **orb** (github.com/paulmach/orb) - Geospatial operations
- **testify** - Testing framework (optional per constitution)
- **golang-migrate** - Database migrations

**Development Considerations**:
- **Hexagonal Architecture** - Maintain ports/adapters separation
- **Environment Configuration** - No hardcoded values
- **Security-First** - Input validation at all boundaries

## Implementation Priority

### Phase 1 Implementation Order:
1. **Core Domain** - Damaged road entity and business logic
2. **Geospatial Validation** - Indonesian boundary checking
3. **Photo URL Validation** - Secure validation with SSRF protection
4. **API Endpoints** - RESTful handlers with Swagger annotations
5. **Database Layer** - PostgreSQL with proper indexing
6. **Authentication** - JWT integration for logged-in users

### Testing Strategy (Optional per Constitution):
- **Unit Tests** - Core business logic validation
- **Integration Tests** - Database and external service integration
- **API Tests** - End-to-end endpoint testing

## Risk Mitigation

### Identified Risks:
1. **Data Source Availability** - Indonesian government data availability
   - **Mitigation**: Local caching with quarterly updates
2. **Network Reliability** - Indonesian network infrastructure
   - **Mitigation**: Optimistic UI with retry mechanisms
3. **Scale Performance** - Nationwide deployment challenges
   - **Mitigation**: Phased scaling with monitoring
4. **Security Vulnerabilities** - Photo URL validation risks
   - **Mitigation**: Multi-layered security approach

### Monitoring Requirements:
- **Response Time Monitoring** - P95 < 2 seconds
- **Error Rate Tracking** - < 1% error rate
- **Security Monitoring** - Failed validation attempts
- **Resource Usage** - Memory, CPU, database connections