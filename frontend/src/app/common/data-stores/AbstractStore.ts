import type { Signal } from '@angular/core';
import { computed, effect, signal } from '@angular/core';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import type { Observable } from 'rxjs';
import type { IFilter } from './IFilter';

export interface StorageOptions<T> {
  key: string;
  isItem: (value: unknown) => value is T;
}

export abstract class AbstractStore<T> {
  protected readonly items = signal<readonly T[]>([]);

  constructor(
    source: Observable<T>,
    private readonly storage: StorageOptions<T>,
  ) {
    this.items.set(this.loadItems());
    effect(() => this.saveItems(this.items()));
    source.pipe(takeUntilDestroyed()).subscribe((item) => this.add(item));
  }

  getAll(): Signal<readonly T[]> {
    return this.items.asReadonly();
  }

  get(filter: IFilter<T>): Signal<readonly T[]> {
    return computed(() => this.items().filter((item) => filter.matches(item)));
  }

  clear(): void {
    this.items.set([]);
  }

  protected add(item: T): void {
    this.items.update((all) => [...all, item]);
  }

  private loadItems(): T[] {
    try {
      const parsed: unknown = JSON.parse(sessionStorage.getItem(this.storage.key) ?? '[]');
      return Array.isArray(parsed) ? parsed.filter(this.storage.isItem) : [];
    } catch {
      return [];
    }
  }

  private saveItems(items: readonly T[]): void {
    try {
      sessionStorage.setItem(this.storage.key, JSON.stringify(items));
    } catch (err) {
      console.warn(`Could not save items to "${this.storage.key}":`, err);
    }
  }
}
