export default class WrongNameCase {
  execute(service: { audit(p: string): void }) {
    service.audit('x');
  }
}
