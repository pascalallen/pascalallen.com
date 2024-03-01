export interface JsonMap {
  [member: string]: string | number | boolean | null | JsonArray | JsonMap;
}
export type JsonArray = Array<string | number | boolean | null | JsonArray | JsonMap>;
export type Json = JsonMap | JsonArray | string | number | boolean | null;

export function cloneJson<T extends Json>(data: T) {
  return <T>JSON.parse(JSON.stringify(data));
}
