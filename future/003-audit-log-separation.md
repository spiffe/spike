     audit log separation (ADR-0027)
      also audit thigns to do:
      1. Make audit entries immutable - Have handlers return audit details rather than mutating the entry.
      2. Distinguish audit event types - Add an EventType field (lifecycle vs operation) to differentiate enter/exit from actual operations.
      3. Populate or remove unused fields - Either extract SPIFFE ID/UserID/SessionID everywhere, or remove them until you're ready to implement.
      4. Fix Resource semantics - Resource should be the actual entity (secret path, policy name), not query params. Maybe add a separate QueryParams field if needed.
      5. Consider structured event nesting - Something like:
      type AuditEvent struct {
      TrailID   string
      RequestEvent  AuditEntry  // enter/exit
      Operations []AuditEntry  // what happened inside
      }
      6. Add audit configuration - Sampling rates, verbosity levels, field inclusion/exclusion for different deployment scenarios.