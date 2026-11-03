import { bind, waive } from '@/speclink.ts';
import WaivedUseCase from '@/domain/usecases/WaivedUseCase.ts';

bind(WaivedUseCase, waive('K5-UC-AUTHZ', 'open to everybody who is signed in'));
