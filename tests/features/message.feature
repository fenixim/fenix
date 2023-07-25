Feature: Sending messages
  Communicate with other users by sending messages

  Scenario: No connection
    Given the server can't be reached
    When I start the app
    Then it will say "failed to connect"

  Scenario: Seeing messages
    Given I have a connection
    When another user sends a message
    Then I should see the message

  Scenario: Sending messages
    Given I have a connection
    When I send a message
    Then everyone (including myself) should see it

  Scenario: Message order
    Given I have a connection
    When multiple users (including myself) send a message at the same time
    Then I should see them in the order the server recieved them
