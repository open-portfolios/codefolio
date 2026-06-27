# Command

Encapsulate a request as an object, letting you parameterize clients with different requests.

## Intent

- Encapsulate a request as an object
- Parameterize objects with different requests
- Support undoable operations
- Queue or log requests

## Implementation

```java
public interface Command {
    void execute();
}

public class Light {
    public void on() {
        System.out.println("Light is on");
    }
    
    public void off() {
        System.out.println("Light is off");
    }
}

public class LightOnCommand implements Command {
    private Light light;
    
    public LightOnCommand(Light light) {
        this.light = light;
    }
    
    @Override
    public void execute() {
        light.on();
    }
}

public class LightOffCommand implements Command {
    private Light light;
    
    public LightOffCommand(Light light) {
        this.light = light;
    }
    
    @Override
    public void execute() {
        light.off();
    }
}

public class RemoteControl {
    private Command command;
    
    public void setCommand(Command command) {
        this.command = command;
    }
    
    public void pressButton() {
        command.execute();
    }
}

// Client code
public class Main {
    public static void main(String[] args) {
        Light light = new Light();
        Command onCommand = new LightOnCommand(light);
        Command offCommand = new LightOffCommand(light);
        
        RemoteControl remote = new RemoteControl();
        
        remote.setCommand(onCommand);
        remote.pressButton();
        
        remote.setCommand(offCommand);
        remote.pressButton();
    }
}
```

## When to Use

- When you want to parameterize objects with actions
- When you want to queue, specify, and execute requests at different times
- When you want to support undo operations
- When you want to support logging changes
