// Strategy for selecting items from a store. Each implementation decides
// which items of type T match, so components can ask a store for exactly
// the items they need.
export interface IFilter<T> {
  matches(item: T): boolean;
}
