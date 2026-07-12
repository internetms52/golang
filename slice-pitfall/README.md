# Go Slice Append 的記憶體共用陷阱

紀錄一下 Go 裡面 `append` 的一個經典坑：**slice 傳進函式後，`append` 可能會、也可能不會共用原本的底層記憶體，取決於 `cap` 夠不夠**。

## 背景知識：Slice 的真面目

Slice 不是陣列本身，而是一個小型 struct（header）：

```go
type slice struct {
    ptr *T   // 指向底層陣列的指標
    len int  // 目前長度
    cap int  // 目前容量
}
```

Go 函式參數是 call by value，所以傳進函式的 slice，複製的是這個 **header**（`ptr`/`len`/`cap` 三個欄位），不是底層陣列本身。

## 核心結論

`append(s, x)` 會不會分配新的底層陣列，只看一個條件：

| 條件 | 行為 | 結果 |
|---|---|---|
| `len(s) < cap(s)`（還有空間） | 直接寫進原本的底層陣列 | 新舊 slice **共用同一塊記憶體** |
| `len(s) == cap(s)`（滿了） | 配置新的、更大的底層陣列，複製舊資料過去 | 新舊 slice **各自獨立** |

`len`/`cap` 這兩個欄位是**屬於每個變數自己的**，不是記憶體本身的屬性。所以：

- 對「既有 index」的資料寫入（`s[i] = x`）→ 只要共用底層陣列，**一定會**互相影響
- 對「新增元素」（append 導致 `len` 增加）→ 只有**被賦值的那個變數**的 `len` 會更新，其他共用同底層陣列的變數看不到新元素（除非它們的 `len` 剛好也涵蓋到那裡）

這就是為什麼會出現一個「有缺陷的 side effect」：資料改了，但外部變數的 `len` 沒有同步更新，導致新增的元素憑空消失，但既有資料的修改卻悄悄外洩。

**執行結果：**

```
3
[100 0 0]
4
[100 0 0 999]
```

### 為什麼會這樣？

`runWithSideEffect`：

1. `original := make([]int, 3, 5)` → 底層陣列 `[0,0,0,_,_]`，`original.len=3, cap=5`
2. 呼叫函式時 `s` 是 `original` header 的拷貝，`ptr` 相同（共用底層陣列）
3. `s = append(s, 999)`：cap 夠，直接寫入底層陣列 index 3，**只有函式內部的 `s.len` 變成 4**
4. `s[0] = 100`：直接改寫共用底層陣列的 index 0
5. 函式結束，`original` 這個變數從頭到尾沒被賦值過，`len` 仍是 3；但底層陣列已經被改成 `[100,0,0,999,_]`
6. 印出來只看 `len` 範圍內的資料：`[100 0 0]`，999 存在但看不到

`runWithoutSideEffect`：

1. 先 `make` + `copy()`，`duplicate` 是全新、獨立的底層陣列
2. 之後不管 `append` 或改值，都不會影響 `original`
3. 因為**有 `return` 且呼叫端有重新賦值** `original = modifyWithoutSideEffect(original)`，`original` 的 header（含 `len`）才會被正確更新

## 不是只有 append 有這個坑

任何「共用底層陣列」的操作，只要是寫入性質，都有一樣的問題：

- 索引賦值 `s[i] = x`
- 切片語法 `s[low:high]` 切出來的子 slice
- `copy(dst, src)` 的目標 `dst`
- `sort.Slice()` / `sort.Ints()` 等 in-place 排序
- 把 slice 傳進函式，內部直接用 index 修改

**不會**有這個坑的：

- 陣列（`[N]T`，非 slice）：value type，傳遞時整個複製
- 已經 `copy()` 過的獨立 slice
- 只讀取、不寫入的操作

## 因應策略

### 心法一：三索引 slicing（輕量、只防 append 誤傳染，不防直接改值）

```go
sub := original[low:high:high] // 讓 max == high，強制 cap == len
sub = append(sub, x)           // cap 不夠，一定觸發新記憶體配置，不會動到 original
```

適合「切一段來用，但不確定會不會 append」的情境。**注意：這只防止 append 造成的污染，不能防止直接改值（`sub[0]=x` 還是會動到共用記憶體）。**

### 心法二：`copy()` 出一份完全獨立的記憶體（保險、通用）

```go
dup := make([]T, len(s))
copy(dup, s)
// 之後對 dup 做任何操作都不會影響 s
```

只要函式裡會對 slice 做**任何寫入操作**（改值、append、排序……），且不希望外洩回呼叫端，就在一開始 copy。單純讀取則不需要。

### 最終判斷準則

> 這個 method 會不會修改 slice？
> - 不會 → 直接用，不用 copy
> - 會，且只是 append（增加長度）→ 讓 `append` 自然回傳新 slice 即可，呼叫端接住重新賦值
> - 會，且要改既有內容（索引賦值/排序）並且不想外洩 → 先 `make` + `copy()` 再操作，最後回傳新 slice

## 一句話總結

> Go 為了效能，讓 slice 保持「輕量 header + 共用底層陣列」的設計，但這個設計天生有「resize 資訊（len/cap）不會自動同步」的副作用。解法不是修補這個副作用，而是遵守一個紀律：**只要函式會修改 slice，就一律回傳新的 slice，呼叫端自己接住並重新賦值**，不要依賴函式內部偷偷 mutate。
