export default class ShapeUseCase {
  execute(service: { audit(p: string): void }) {
    service.audit('x');
  }

  other() {}
}
