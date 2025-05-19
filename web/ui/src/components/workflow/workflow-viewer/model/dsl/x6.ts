import yaml from 'yaml';
import { DslTypeEnum } from '@/api/dto/enumeration';

export function parse(dsl: string | undefined): {
  dslType: DslTypeEnum;
  asyncTaskRefs: string[];
  data: string;
} {
  if (!dsl ) {
    return { dslType: DslTypeEnum.PIPELINE, asyncTaskRefs: [], data: '' };
  }

  const { workflow, pipeline, 'raw-data': rawData } = yaml.parse(dsl);

  let data: string;
  data = rawData;

  return {
    dslType: workflow ? DslTypeEnum.WORKFLOW : DslTypeEnum.PIPELINE,
    asyncTaskRefs: Object.keys(workflow || pipeline),
    data,
  };
}