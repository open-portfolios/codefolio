# Iterator

Provide a way to access elements of a collection sequentially without exposing its underlying representation.

## Intent

- Access elements of a collection sequentially
- Decouple iteration from the collection
- Support multiple traversals simultaneously

## Implementation

```java
public interface Iterator<T> {
    boolean hasNext();
    T next();
}

public class ArrayIterator<T> implements Iterator<T> {
    private T[] array;
    private int position = 0;
    
    public ArrayIterator(T[] array) {
        this.array = array;
    }
    
    @Override
    public boolean hasNext() {
        return position < array.length;
    }
    
    @Override
    public T next() {
        return array[position++];
    }
}

public class ListIterator<T> implements Iterator<T> {
    private List<T> list;
    private int position = 0;
    
    public ListIterator(List<T> list) {
        this.list = list;
    }
    
    @Override
    public boolean hasNext() {
        return position < list.size();
    }
    
    @Override
    public T next() {
        return list.get(position++);
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        String[] array = {"A", "B", "C"};
        Iterator<String> arrayIterator = new ArrayIterator<>(array);
        
        while (arrayIterator.hasNext()) {
            System.out.println(arrayIterator.next());
        }
        
        List<String> list = Arrays.asList("X", "Y", "Z");
        Iterator<String> listIterator = new ListIterator<>(list);
        
        while (listIterator.hasNext()) {
            System.out.println(listIterator.next());
        }
    }
}
```

## When to Use

- When you want to access a collection's elements without exposing its internal structure
- When you want to provide multiple traversal methods for the same collection
- When you want to decouple collection logic from iteration logic
