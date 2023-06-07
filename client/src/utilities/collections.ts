import _ from 'lodash';
import { Json } from '@domain/types/Json';

export const listToMap = (array: string[]): { [key: string]: string } => {
  const map: { [key: string]: string } = {};

  array.forEach(item => {
    map[item] = item;
  });

  return map;
};

type QueryParamsObject = {
  [key: string]: Json;
};

export const removeEmptyKeys = (params: QueryParamsObject): QueryParamsObject => {
  _.forEach(params, (value, key) => {
    if (
      _.isNil(value) ||
      _.isUndefined(value) ||
      _.isNaN(value) ||
      (_.isArray(value) && _.isEmpty(value)) ||
      value === ''
    ) {
      delete params[key];
    }
    if (_.isObject(value) || _.isArray(value)) {
      _.forEach(params, (value, key) => {
        if (
          _.isNil(value) ||
          _.isUndefined(value) ||
          _.isNaN(value) ||
          (_.isArray(value) && _.isEmpty(value)) ||
          value === ''
        ) {
          delete params[key];
        }
      });
    }
  });

  return params;
};
