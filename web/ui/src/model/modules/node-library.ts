import { INodeVo } from '@/api/dto/node-library';
import { Mutable } from '@/utils/lib';

export interface INode<T> {
  total?: number;
  list: T[];
}
