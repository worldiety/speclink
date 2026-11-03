/**
 * The speclink vocabulary.
 *
 * speclink reads these calls from the source and never runs them; the functions exist so that
 * the compiler checks every reference a requirement or a binding makes. Requirements live in
 * requirements/<kind>/…/<ID>.req.ts, bindings beside the code in <module>.annotation.ts.
 */

export enum Kind {
  Functional = 'Functional',
  NonFunctional = 'NonFunctional',
  Constraint = 'Constraint',
  Decision = 'Decision'
}

export enum Status {
  Normative = 'Normative',
  Abstract = 'Abstract',
  Planned = 'Planned',
  OutOfScope = 'OutOfScope',
  Informative = 'Informative',
  Superseded = 'Superseded'
}

export enum Discipline {
  Business = 'Business',
  Technical = 'Technical',
  Mixed = 'Mixed'
}

export enum Role {
  Mockup = 'Mockup',
  Scribble = 'Scribble',
  Diagram = 'Diagram',
  AcceptanceCriteria = 'AcceptanceCriteria',
  Protocol = 'Protocol',
  Document = 'Document'
}

export interface Source {
  doc?: string;
  anchor?: string;
  extern?: string;
  note?: string;
}

export interface Attachment {
  path: string;
  role: Role;
  note?: string;
}

export interface Requirement {
  id: string;
  kind: Kind;
  status: Status;
  discipline?: Discipline;
  title?: string;
  text: string;
  detail?: string;
  rationale?: string;
  consequences?: string;
  derivedFrom?: Requirement[];
  supersedes?: Requirement[];
  sources?: Source[];
  attachments?: Attachment[];
}

export interface Assertion {
  readonly assertion: string;
}

/**
 * Declares a requirement: export const RTsvActionOverview = requirement({ … });
 */
export function requirement(declaration: Requirement): Requirement {
  return declaration;
}

/**
 * States that the bound construct was written for these requirements.
 */
export function satisfies(..._requirements: Requirement[]): Assertion {
  return { assertion: 'satisfies' };
}

/**
 * Suspends one rule for the bound construct. The reason is mandatory and appears in the report.
 */
export function waive(_rule: string, _reason: string): Assertion {
  return { assertion: 'waive' };
}

/**
 * Says why the bound construct is shaped the way it is.
 */
export function rationale(_text: string): Assertion {
  return { assertion: 'rationale' };
}

/**
 * Binds assertions to a class: bind(LoadActionOverviewUseCase, satisfies(RTsvActionOverview));
 */
// eslint-disable-next-line @typescript-eslint/no-explicit-any
export function bind(
  _target: abstract new (...args: any[]) => unknown,
  ..._assertions: Assertion[]
) {}

/**
 * Binds assertions to the module the annotation file sits beside.
 */
export function bindModule(..._assertions: Assertion[]) {}

/**
 * Claims, inside a test, that the test demonstrates these requirements.
 */
export function verified(..._requirements: Requirement[]) {}
