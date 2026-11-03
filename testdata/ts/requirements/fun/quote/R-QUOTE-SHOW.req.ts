import { requirement, Kind, Status } from '../../../src/speclink.ts';
import { RDecHexagonal } from '../../dec/R-DEC-HEXAGONAL.req.ts';

export const RQuoteShow = requirement({
  id: 'R-QUOTE-SHOW',
  kind: Kind.Functional,
  status: Status.Normative,
  title: 'Show a quote',
  text:
    'A quote MUST be shown with its number ' +
    'and its amount.',
  derivedFrom: [RDecHexagonal],
  sources: [{ doc: 'requirements/_sources/quotes.md', anchor: '1-anzeige' }]
});
