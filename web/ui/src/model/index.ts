/**
 * 滚动偏移量
 */
export interface IScrollOffset {
  left: number;
  top: number;
}

/**
 * vuex根状态
 */
export interface IRootState {
  version: "";
  thirdPartyType: string;
  authMode: string;
  parameterTypes: string[];
  fromRoute: {
    path: string;
    fullPath: string;
  };
  scrollbarOffset: {
    [fullPath: string]: IScrollOffset;
  };
}
