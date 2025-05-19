export interface IErrorMessageVo extends Readonly<{
  statusCode: number;
  timestamp: Date;
  message: string;
  description: string;
}> {
}

/**
 * 抽象值对象
 */
export interface IPageDto extends Readonly<{
  /**
   * 页码
   */
  pageNum: number;

  /**
   * 每页个数
   */
  pageSize: number;
}> {
}

/**
 * 抽象值对象
 */
export interface BaseVo extends Readonly<{
  /**
   * 创建时间
   */
  createdTime: string;

  /**
   * 最后修改时间
   */
  modifiedTime: string;
}> {
}

/**
 * 分页值对象
 */
export interface IPageVo<T> extends Readonly<{
  /**
   * 总个数
   */
  total: number;

  /**
   * 总页数
   */
  pages: number;

  /**
   * 当前页码
   */
  pageNum: number;

  /**
   * 数据
   */
  list: T[];
}> {
}