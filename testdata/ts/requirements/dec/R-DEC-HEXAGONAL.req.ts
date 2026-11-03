import { requirement, Kind, Status } from '../../src/speclink.ts';

export const RDecHexagonal = requirement({
  id: 'R-DEC-HEXAGONAL',
  kind: Kind.Decision,
  status: Status.Normative,
  title: 'Ports and adapters',
  text: 'The domain MUST reach the outside only through ports.',
  rationale: 'A backend that changes must not change the rules of the business.',
  sources: [{ extern: 'Team decision of 2026-10-01' }]
});
