# Observer

Define a one-to-many dependency between objects so that when one object changes state, all its dependents are notified.

## Intent

- Define a one-to-many dependency between objects
- Notify dependents when an object changes state
- Loose coupling between subject and observers

## Implementation

```java
public interface Observer {
    void update(String message);
}

public interface Subject {
    void attach(Observer observer);
    void detach(Observer observer);
    void notifyObservers();
}

public class NewsAgency implements Subject {
    private List<Observer> observers = new ArrayList<>();
    private String news;
    
    @Override
    public void attach(Observer observer) {
        observers.add(observer);
    }
    
    @Override
    public void detach(Observer observer) {
        observers.remove(observer);
    }
    
    @Override
    public void notifyObservers() {
        for (Observer observer : observers) {
            observer.update(news);
        }
    }
    
    public void setNews(String news) {
        this.news = news;
        notifyObservers();
    }
}

public class NewsChannel implements Observer {
    private String name;
    
    public NewsChannel(String name) {
        this.name = name;
    }
    
    @Override
    public void update(String news) {
        System.out.println(name + " received: " + news);
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        NewsAgency agency = new NewsAgency();
        
        NewsChannel channel1 = new NewsChannel("Channel 1");
        NewsChannel channel2 = new NewsChannel("Channel 2");
        
        agency.attach(channel1);
        agency.attach(channel2);
        
        agency.setNews("Breaking news!");
    }
}
```

## When to Use

- When changes to one object require changing others, and you don't know how many objects need to change
- When an object should notify other objects without making assumptions about these objects
- When you need loose coupling between objects
