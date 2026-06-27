# Mediator

Define an object that encapsulates how a set of objects interact. Mediator promotes loose coupling by keeping objects from referring to each other explicitly.

## Intent

- Encapsulate interaction between objects
- Decouple objects from knowing about each other
- Centralize control logic

## Implementation

```java
public interface ChatMediator {
    void sendMessage(String message, User sender);
    void addUser(User user);
}

public class ChatRoom implements ChatMediator {
    private List<User> users = new ArrayList<>();
    
    @Override
    public void sendMessage(String message, User sender) {
        for (User user : users) {
            if (user != sender) {
                user.receiveMessage(message, sender.getName());
            }
        }
    }
    
    @Override
    public void addUser(User user) {
        users.add(user);
    }
}

public class User {
    private String name;
    private ChatMediator mediator;
    
    public User(String name, ChatMediator mediator) {
        this.name = name;
        this.mediator = mediator;
    }
    
    public void send(String message) {
        System.out.println(name + " sends: " + message);
        mediator.sendMessage(message, this);
    }
    
    public void receiveMessage(String message, String sender) {
        System.out.println(name + " received from " + sender + ": " + message);
    }
    
    public String getName() {
        return name;
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        ChatMediator mediator = new ChatRoom();
        
        User user1 = new User("Alice", mediator);
        User user2 = new User("Bob", mediator);
        User user3 = new User("Charlie", mediator);
        
        mediator.addUser(user1);
        mediator.addUser(user2);
        mediator.addUser(user3);
        
        user1.send("Hello everyone!");
    }
}
```

## When to Use

- When you have a set of objects with complex communication
- When you want to reduce coupling between objects
- When you want to centralize complex interaction logic
