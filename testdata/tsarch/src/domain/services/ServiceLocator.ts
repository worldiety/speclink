import ThingService from './ThingService';

export default class ServiceLocator {
  private static thingService: ThingService;

  public static init() {}

  public static getThingService(): ThingService {
    return this.thingService;
  }
}
