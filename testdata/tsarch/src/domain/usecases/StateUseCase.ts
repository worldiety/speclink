let calls = 0;

export default class StateUseCase {
  execute(service: { audit(p: string): void }) {
    service.audit('x');
    calls++;
    return calls;
  }
}
