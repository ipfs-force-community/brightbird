import {
  IProjectGroupCreatingDto,
  IProjectGroupEditingDto,
} from '@/api/dto/testflow-group';

import {
  IChangeTestflowGroupDto
} from '@/api/dto/project'
import { Mutable } from '@/utils/lib';

export interface IProjectGroupCreateFrom
  extends Mutable<IProjectGroupCreatingDto> {
}

export interface IProjectGroupEditFrom
  extends Mutable<IProjectGroupEditingDto> {
}

export interface IProjectGroupAddingForm
  extends Mutable<IChangeTestflowGroupDto> {
}

export interface IState {
  [key: string]: boolean
}
