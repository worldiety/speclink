import { bind, satisfies } from '@/speclink.ts';
import ShowQuoteUseCase from '@/domain/usecases/ShowQuoteUseCase.ts';
import { RQuoteShow } from '../../../requirements/fun/quote/R-QUOTE-SHOW.req.ts';
import { RDecHexagonal } from '../../../requirements/dec/R-DEC-HEXAGONAL.req.ts';

bind(ShowQuoteUseCase, satisfies(RQuoteShow, RDecHexagonal));
