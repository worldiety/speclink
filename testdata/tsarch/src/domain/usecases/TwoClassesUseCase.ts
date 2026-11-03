export default class TwoClassesUseCase {
  execute(service: { audit(p: string): void }) {
    service.audit('x');
  }
}

export class SecondClass {}
