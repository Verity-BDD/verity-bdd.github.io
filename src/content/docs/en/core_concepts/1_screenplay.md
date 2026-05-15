---
title: Screenplay Pattern
description: What is Screenplay Pattern?
sidebar:
  order: 1
---

## Screenplay Pattern

The **Screenplay Pattern** is an innovative, [user-centred](https://en.wikipedia.org/wiki/User-centered_design) approach
to writing high-quality automated acceptance tests.
It steers your team towards effectively using **layers of abstraction**,
helps your test scenarios capture the **business vocabulary** of **your domain**,
and encourages good testing and software engineering habits.

Focusing on actors and their goals and incorporating your domain language into test scenarios
improves **team collaboration** and **alignment**, enabling technical and business stakeholders to understand
and readily contribute to the test automation process.

### The design principle

The design principle behind the Screenplay Pattern is simple but might forever change the way you look at test automation:

:::tip[Remember]
**Automated acceptance tests** should use **your domain language** to clearly express what **activities** the **actors**
interacting with **your system** need to perform in order to accomplish **their goals**.
:::

Applying this design principle to your automated tests has a number of positive implications:
- Expressing your test scenarios in **your domain language** makes them easier to understand and accessible to a wider audience
- Focusing on **actors** and **their goals** makes it easy to correlate any test failures with the actual business impact
- Modelling actor workflows using sequences of business-focused, reusable **activities** reduces code duplication, improves flexibility of your test codebase,
and means that your team can quickly compose new test scenarios from existing steps

> To communicate effectively, the code must be based on the same language used to write the requirements—the same language that the developers speak with each other and with domain experts.
>
> ― _Eric Evans, ["Domain-Driven Design: Tackling Complexity in the Heart of Software"](https://amzn.to/3lKVpFv)_

### The five elements of the Screenplay Pattern

The **Screenplay Pattern** uses the [system metaphor](https://wiki.c2.com/?SystemMetaphor) of a **stage performance**,
helping you model each test scenario like a little [screenplay](https://en.wikipedia.org/wiki/Screenplay)
describing how the actors should go about performing their activities
while interacting with the [system under test](http://xunitpatterns.com/SUT.html).

Following the Screenplay Pattern helps you capture:
- **Who** the actors interacting with your system are
- **Why** they interact with your system
- **What** they need to do to accomplish their goals
- **How** exactly they would go about that

The five building blocks of the Screenplay Pattern are:
- **[Actors](#actors)**, who represent **people** and **external systems** interacting with the system under test
- **[Abilities](#abilities)**, that act as **thin wrappers** around any **integration libraries** required to interact with the system under test
- **[Interactions](#interactions)**, which represent the **low-level activities** an actor can perform using a given interface
- **[Tasks](#tasks)**, used to model **sequences of activities** as meaningful steps of a business workflow in your domain
- **[Questions](#questions)**, used to **retrieve information** from the system under test and the test execution environment

### Screenplay Pattern with Verity BDD

The best way to illustrate the Screenplay Pattern is through a practical example, so assume for a moment that we're writing a test scenario
for an online shop. The shop has a REST API that lets us configure its product catalogue with some test data,
and a web storefront that lets customers find the products they need and make a purchase.

We'll create a test scenario that uses two actors: one to set up the test data via the REST API, and one to verify the results.

:::tip[Verity BDD Project Templates]
To follow along with the coding, get one of the [**Verity BDD Project Templates**](/handbook/project-templates/) as they come with everything you need to get started with Verity BDD.
:::

#### Actors

A test scenario following the Screenplay Pattern has one or multiple **actors**
representing people and external systems interacting with the system under test and playing specific roles.

:::tip[The role of an actor]
Just like the five core elements of the Screenplay Pattern, the term "role" comes from the system metaphor of a **stage performance**.
It should be interpreted as the role a given actor plays in the performance,
the role they play in the system. Some good examples of roles include "a doctor", "a trader", or "a writer".

While a "role" might imply the _permissions_ a given actor has in the system they interact with (e.g. a "writer" can write articles,
but only a "publisher" publishes articles), this is not a mechanism to _prevent_ the actor from performing activities inconsistent with their role.

In particular, Verity BDD will not prevent you from writing scenarios where a "writer" tries to impersonate a "publisher"
and publish an article. If it did, you would not be able to test if your system correctly implemented its access control mechanisms!
:::

Our example scenario could have two actors, who we'll call:
- Apisitt, responsible for setting up test data using the REST API
- Wendy, verifying the results via the REST API

##### Instantiating and retrieving actors

With Verity BDD, you instantiate actors via the `VerityTest` context created at the start of each test:

```go title="online_shop_test.go"
import (
    "context"
    "testing"

    verity "github.com/nchursin/verity-bdd"
)

func TestOnlineShop(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    test.ActorCalled("Apisitt") // Actor{name: "Apisitt"}
    test.ActorCalled("Wendy")   // Actor{name: "Wendy"}
}
```

Note that **every Verity BDD actor is uniquely identified by their name**.
The first time you call `test.ActorCalled("Wendy")`, Verity BDD instantiates a new `Actor`
and stores a reference to it internally under the name you gave it.
This way, whenever you call `test.ActorCalled("Wendy")` **within the same test** again, you'll get the same actor instance back:

```go title="online_shop_test.go"
func TestActorIdentity(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    wendy1 := test.ActorCalled("Wendy") // first call — actor created
    wendy2 := test.ActorCalled("Wendy") // second call — same instance returned

    // wendy1 and wendy2 are the same Actor instance
}
```

To avoid typos and repetition when instantiating and retrieving actors in your test scenarios,
you might want to consider using constants to store actor names:

```go title="online_shop_test.go"
const (
    actorApisitt = "Apisitt, the test data manager"
    actorWendy   = "Wendy, the customer"
)

func TestOnlineShop(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    test.ActorCalled(actorApisitt) // Actor{name: "Apisitt, the test data manager"}
    test.ActorCalled(actorWendy)   // Actor{name: "Wendy, the customer"}
}
```

##### Using actors in Go tests

While you could use Verity BDD and `test.ActorCalled()` as part of any regular Go program,
you'll typically use it with Go's standard `testing` package:

```go title="online_shop_test.go"
import (
    "context"
    "testing"

    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_abilities/api"
)

func TestOnlineShop(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    test.ActorCalled("Apisitt") // Actor{name: "Apisitt"}
    test.ActorCalled("Wendy")   // Actor{name: "Wendy"}
}
```

`VerityTest` automatically cleans up all actors when the test finishes — no manual teardown needed.

##### Using default abilities

Since giving every actor the same base ability (such as the API base URL) is a common need,
Verity BDD lets you configure **default abilities** via `Scene.DefaultAbilities`.
Every actor created in the test will receive these abilities automatically:

```go title="online_shop_test.go"
func TestOnlineShop(t *testing.T) {
    test := verity.NewVerityTest(t, verity.Scene{
        DefaultAbilities: []verity.DefaultAbilityFactory{
            func(_ string) verity.Ability {
                return api.CallAnApiAt("https://api.example.org/")
            },
        },
    })

    apisitt := test.ActorCalled("Apisitt") // already has CallAnApiAt ability
    wendy := test.ActorCalled("Wendy")     // already has CallAnApiAt ability
    _ = apisitt
    _ = wendy
}
```

To retrieve an actor's ability, use `verity.AbilityOf[T]`:

```go title="online_shop_test.go"
ability, err := verity.AbilityOf[*api.CallAnAPI](apisitt)
if err != nil {
    t.Fatalf("actor lacks API ability: %v", err)
}
_ = ability
```

Learn more about:
- [Abilities](#abilities)
- [API testing](/guides/1_getting-started/)

#### Abilities

Actors have **abilities** that enable them to interact with the various interfaces
of the system under test and the test execution environment.

From the technical perspective, **abilities** act as **wrappers** around any **integration libraries** required
to communicate with the external interfaces of the system under test.

:::note[Did you know?]
The word "screen" in "screenplay" has nothing to do with the _computer screen_.
On the contrary, the Screenplay Pattern is a **general method** of modelling acceptance tests interacting with _any_
external interface of your system. In fact, Verity BDD implementation of the Screenplay Pattern can help you
break free from UI-only-based testing!
:::

To allow Apisitt to interact with a REST API, we give him the ability to `CallAnApiAt`:

```go title="online_shop_test.go"
import (
    "context"
    "testing"

    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_abilities/api"
)

func TestOnlineShop(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    apisitt := test.ActorCalled("Apisitt").
        WhoCan(api.CallAnApiAt("https://api.example.org/"))

    wendy := test.ActorCalled("Wendy").
        WhoCan(api.CallAnApiAt("https://api.example.org/"))

    _ = apisitt
    _ = wendy
}
```

Learn more about:
- [Abilities](/api/core/class/Ability)
- [API testing](/guides/1_getting-started/)

#### Interactions

Abilities enable actors to perform **interactions** with the system under test.
**Interactions** are **command objects** that instruct an actor how to use their abilities to perform the given activity.
Most interactions you will need are already provided by Verity BDD,
and you can easily create new ones if you'd like to.

To instruct an actor to attempt to perform a sequence of interactions, use the `AttemptsTo` method.

Here, we instruct Apisitt to use `api.SendPostRequest` to set up some test data for our test scenario:

```go title="online_shop_test.go"
import (
    "context"
    "testing"

    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_abilities/api"
)

type Product struct {
    Name  string `json:"name"`
    Price string `json:"price"`
}

func TestOnlineShop(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    apisitt := test.ActorCalled("Apisitt").
        WhoCan(api.CallAnApiAt("https://api.example.org/"))

    apisitt.AttemptsTo(                        // actor attempts to perform interactions
        api.SendPostRequest("/products").With(  // interactions are command objects
            []Product{                         // that instruct actors how to use abilities
                {Name: "Apples", Price: "£2.50"},
            },
        ),
    )
}
```

In the same manner, Wendy can use `api.SendGetRequest` to verify the catalogue:

```go title="online_shop_test.go"
wendy := test.ActorCalled("Wendy").
    WhoCan(api.CallAnApiAt("https://api.example.org/"))

wendy.AttemptsTo(
    api.SendGetRequest("/products"), // consistent API, same as any other interaction
)
```

If you wanted to implement a custom interaction yourself, you can use `verity.Do`:

```go title="interactions.go"
import (
    "context"
    "fmt"

    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_abilities/api"
)

func GetProducts() verity.Activity {
    return verity.Do(
        "#actor retrieves the product catalogue",
        func(ctx context.Context, actor verity.Actor) error {
            ability, err := verity.AbilityOf[*api.CallAnAPI](actor)
            if err != nil {
                return fmt.Errorf("actor needs API ability: %w", err)
            }
            _ = ability
            // perform the request using the ability
            return nil
        },
    )
}
```

Learn more about:
- [Actors](/api/core/class/Actor)
- [Interactions](/api/core/class/Interaction)

#### Questions

Apart from enabling interactions, abilities also enable actors to answer **questions**
about the state of the system under test and the test execution environment.
More specifically, **questions** instruct actors how to use their abilities to **retrieve information** and provide a way to **parameterise activities**.

When Apisitt sends a `PostRequest`, the response is stored in his API ability.
To assert on it, we use `api.LastResponseStatus{}` — a question about the HTTP status code — together with `ensure.That`:

```go title="online_shop_test.go"
import (
    "context"
    "testing"

    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_abilities/api"
    expectations "github.com/nchursin/verity-bdd/verity_expectations"
    "github.com/nchursin/verity-bdd/verity_expectations/ensure"
)

func TestOnlineShop(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    apisitt := test.ActorCalled("Apisitt").
        WhoCan(api.CallAnApiAt("https://api.example.org/"))

    apisitt.AttemptsTo(
        api.SendPostRequest("/products").With([]Product{
            {Name: "Apples", Price: "£2.50"},
        }),
        ensure.That(
            api.LastResponseStatus{},
            expectations.Equals(201),
        ),
    )
}
```

An excellent proof of the **design consistency** enabled by the Verity BDD Screenplay Pattern is that
no matter what ability an actor uses, the way they answer questions and assert on responses is always the same:

```go title="online_shop_test.go"
apisitt.AttemptsTo(
    api.SendPostRequest("/products").With([]Product{
        {Name: "Apples", Price: "£2.50"},
    }),
    ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
)

wendy.AttemptsTo(
    api.SendGetRequest("/products"),
    ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
    ensure.That(api.LastResponseBody{}, expectations.Contains("Apples")),
)
```

Learn more about:
- [Questions](/api/core/class/Question)
- [Assertions and expectations](/handbook/design/assertions)

#### Tasks

The idea that underpins the Screenplay Pattern is to **capture your domain language** and use your acceptance tests as an opportunity to demonstrate
how actors interacting with your system accomplish their goals.

Conceptually similar to standard Go functions, **tasks** offer an easy way to **associate business meaning** with **sequences of activities**
and turn them into **reusable building blocks** from which your team can assemble test scenarios.

For example, we can use `verity.TaskWhere` to define custom tasks that capture how an actor would set up a product catalogue:

```go title="tasks.go"
import (
    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_abilities/api"
    expectations "github.com/nchursin/verity-bdd/verity_expectations"
    "github.com/nchursin/verity-bdd/verity_expectations/ensure"
)

type Product struct {
    Name  string `json:"name"`
    Price string `json:"price"`
}

func SetupProductCatalogue(products []Product) verity.Activity {
    return verity.TaskWhere("#actor sets up the product catalogue",
        api.SendPostRequest("/products").With(products),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(201)),
    )
}

func VerifyProductCatalogue() verity.Activity {
    return verity.TaskWhere("#actor verifies the product catalogue",
        api.SendGetRequest("/products"),
        ensure.That(api.LastResponseStatus{}, expectations.Equals(200)),
        ensure.That(api.LastResponseBody{}, expectations.Contains("Apples")),
    )
}
```

As you can see, custom tasks like these are easy to read and understand, and can be parameterised and reused across different test scenarios.
Tasks help you capture the domain language, provide a consistent way to structure your test scenarios, and make your test code easier to maintain.

```go title="online_shop_test.go"
import (
    "context"
    "testing"

    verity "github.com/nchursin/verity-bdd"
    "github.com/nchursin/verity-bdd/verity_abilities/api"
)

func TestOnlineShop(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    apisitt := test.ActorCalled("Apisitt").
        WhoCan(api.CallAnApiAt("https://api.example.org/"))

    wendy := test.ActorCalled("Wendy").
        WhoCan(api.CallAnApiAt("https://api.example.org/"))

    apisitt.AttemptsTo(
        SetupProductCatalogue([]Product{
            {Name: "Apples", Price: "£2.50"},
        }),
    )

    wendy.AttemptsTo(
        VerifyProductCatalogue(),
    )
}
```

Learn more about:
- [Actors](/api/core/class/Actor)
- [Tasks](/api/core/class/Task)

### Performing activities at multiple levels

The role of an actor is to perform activities that demonstrate how a goal can be accomplished at the given **level of abstraction**.

:::tip[Remember]
**Actors** represent **people** and **external systems** interacting with the system under test.
:::

For example, we might have an acceptance test that demonstrates how the system under test enables an actor to accomplish the goal of booking a flight.

```go title="flight_booking_test.go"
func TestFlightBooking(t *testing.T) {
    test := verity.NewVerityTest(ctx, verity.Scene{})

    trevor := test.ActorCalled("Trevor").
        WhoCan(api.CallAnApiAt("https://api.airline.org/"))

    trevor.AttemptsTo(
        FindFlight("London", "New York"),   // activities describe the business goal,
        ChooseFlightClass("Economy"),        // not the underlying API calls
        ProvidePaymentDetails(defaultCard),
        ReceiveBookingConfirmation(),
    )
}
```

If you are using a BDD framework like [godog](https://github.com/cucumber/godog), the name of the feature, the goal of the scenario,
as well as the high-level steps necessary to achieve the goal would already be captured in `.feature` files:

```gherkin
Feature: Serenity Airlines flight booking                                       # system feature

  Scenario: traveller books a plane ticket                                      # scenario goal

    Given Trevor finds a flight from 'London' to 'New York'                     # high-level steps
      And he chooses the 'Economy' flight class
     When he provides his payment details
     Then he should receive a booking confirmation
```

In this case, each step definition is mapped to a Verity BDD actor performing one or more activities:

```go title="flight_booking_steps_test.go"
import (
    "github.com/cucumber/godog"
    verity "github.com/nchursin/verity-bdd"
)

func (s *suite) TrevorFindsAFlight(origin, destination string) error {
    return s.trevor.AttemptsTo(
        FindFlight(origin, destination),     // step goal maps to activities
    )
}

func (s *suite) HeChoosesFlightClass(class string) error {
    return s.trevor.AttemptsTo(
        ChooseFlightClass(class),
    )
}
```

:::tip[Remember]
**Actors** demonstrate how to accomplish a **goal** by performing **activities** at multiple **levels of abstraction**.
:::

At the **high levels of abstraction**, e.g. in business-focused acceptance test scenarios,
the vocabulary we use is rooted in the **business domain**, and so are the names we choose for the activities.

For example, an acceptance test scenario might state that for the system to enable the actor to accomplish the **goal** of _booking a plane ticket_,
an actor should be able to successfully perform the following **high-level activities**:
- find an appropriate flight connection,
- choose flight class,
- provide payment details,
- receive booking confirmation.

The names we give functions that produce those activities, such as `FindFlight` or `ChooseFlightClass`,
represent those steps in the business process and are agnostic of the interface through which actors interact with the system under test.

:::note[Model the expected process, not the existing implementation]
When describing an acceptance test at a high level of abstraction,
the way we **name the activities** is focused on representing the steps of the **expected business process**
and **not tied to the implementation** of any specific interface of the system under test.
"Find an appropriate flight connection", "choose flight class", or "provide payment details"
are all good examples of such high-level activity names.

This design approach helps to produce test scenarios that are **easier to read and understand** and to a much **wider audience** than the traditional test scripts.
It also results in two other major advantages:
- once the business process is clearly described in our test scenario, we can often use our acceptance tests to identify **obstacles in user journeys**, or even **highlight errors** and **hidden assumptions** in the **business process** itself
- since we're not tying the implementation to any particular interface, we leave ourselves **more integration options** when it comes to automation.

After all, most business process steps could be accomplished in different ways.
An actor could `FindFlight` by interacting with a web UI, a mobile app, by sending requests to a web service, or even by actually going to the ticket office at the airport!
:::

At the **low level of abstraction**, the vocabulary we use to describe actor's activities
is focused on the **interface** the actor needs to interact with.
Here the goal might be to _use the REST API to search for available flights_. To accomplish it, the actor would need to:
- send a `GET` request to `/flights` with origin and destination parameters
- assert that the response contains at least one available flight

:::tip[Remember]
The core idea behind the Screenplay Pattern is to express the acceptance tests from the perspective of **actors**
playing a certain **role** and attempting to accomplish their **goals** by performing **activities** at **multiple levels** of abstraction.
:::

Of course, most activities fall **somewhere in between** the high and low levels of abstraction.
Furthermore, turns out that **higher-level** activities can be **composed of lower-level activities**,
which themselves could be composed of _**even lower-level activities**_!

If you're familiar with [User Experience Design](https://en.wikipedia.org/wiki/User_experience_design),
you might recognise this style of [functional decomposition](https://en.wikipedia.org/wiki/Functional_decomposition)
from [Hierarchical Task Analysis](https://janmolak.com/user-centred-design-how-a-50-year-old-technique-became-the-key-to-scalable-test-automation-66a658a36555).

The fascinating aspect of looking at your test scenarios as sequences of activities made up of activities,
made up of activities, is that this mental model lends itself perfectly to [functional composition](https://en.wikipedia.org/wiki/Function_composition_(computer_science))
and making _activities_ the primary component of code reuse in Verity BDD.

### Start with Verity BDD Screenplay Pattern

The easiest way to **experience** working with Verity BDD and the Screenplay Pattern is
to **follow the getting started guide** and write your [**first API scenario**](/en/guides/1_getting-started/)!

<!-- :::tip[Try Verity BDD in your browser] -->
<!-- Thanks to [GitHub Codespaces](/handbook/project-templates/#serenityjs-codespaces), -->
<!-- you can follow the tutorial and use any of the [Verity BDD Project Templates](/handbook/project-templates/) right here in your browser, -->
<!-- no local installation required! -->
<!-- ::: -->

### History of the Screenplay Pattern

[Verity BDD](https://github.com/nchursin/verity-bdd) is a Go implementation of the Screenplay Pattern,
but the ideas behind the pattern have been around since 2007 in various forms.

This list is a chronological order of significant events, implementations, and writings related to the evolution of the Screenplay Pattern.

:::note[Credits]
* 2007: <abbr title="Agile Alliance Functional Test Tools">AAFTT</abbr> workshop - [In praise of abstraction](https://developertesting.com/archives/month200710/20071013-In%20Praise%20of%20Abstraction.html) - **Kevin Lawrence** introduces the idea of using the language of interaction designers to model automated tests
* 2007: <abbr title="Agile Alliance Functional Test Tools">AAFTT</abbr> workshop - [**Antony Marcano**](http://antonymarcano.com/) demonstrates the ["Roles, Goals, Tasks, Actions" model](http://antonymarcano.com/blog/2011/03/goals-tasks-action/), which later evolves into the Screenplay Pattern
* 2008: [JNarrate](https://www.slideshare.net/RiverGlide/a-journey-beyond-the-page-object-pattern) - first experimental Java implementation of the "Roles, Goals, Tasks, Actions" model by Antony Marcano and [**Andy Palmer**](https://andypalmer.com/)
* 2011: [Cuke Salad](https://github.com/RiverGlide/CukeSalad/) - Ruby implementation of the "Roles, Goals, Tasks, Actions" model by Antony Marcano
* 2011: [A bit of UCD for BDD & ATDD: Goals -> Tasks -> Actions](http://antonymarcano.com/blog/2011/03/goals-tasks-action/) - blog post by Antony Marcano explaining the motivation behind the "Roles, Goals, Tasks, Actions" model
* 2012: [Screenplay4j](https://bitbucket-archive.softwareheritage.org/projects/te/testingreflections/screenplay4j.html) - first public Java implementation by Antony Marcano and Andy Palmer
* 2012: [User Centred Scenarios: Describing capabilities, not solutions](https://skillsmatter.com/skillscasts/3141-user-centred-scenarios-describing-capabilities-not-solutions) - talk by Antony Marcano and **James Martin**
* 2013: [ScreenplayJVM](https://github.com/screenplay/screenplay-jvm) - Java implementation by Antony Marcano and [**Jan Molak**](https://linkedin.com/in/janmolak)
* 2013: [A journey beyond the page object pattern](https://www.slideshare.net/RiverGlide/a-journey-beyond-the-page-object-pattern) - talk by Antony Marcano, Jan Molak and [**Kostas Mamalis**](https://www.linkedin.com/in/kostasmamalis)
* 2015: [Serenity BDD](http://serenity-bdd.info/) - [**John Ferguson Smart**](https://www.linkedin.com/in/john-ferguson-smart/) and Jan Molak, along with Andy Palmer and Antony Marcano, add native support for the Screenplay Pattern to Serenity BDD, popularising the pattern in the Java testing community
* 2016: [Beyond Page Objects: Next Generation Test Automation with Serenity and the Screenplay Pattern](https://www.infoq.com/articles/Beyond-Page-Objects-Test-Automation-Serenity-Screenplay/) by Andy Palmer, Antony Marcano, John Ferguson Smart, and Jan Molak
* 2016: [Page Objects Refactored: SOLID Steps to the Screenplay/Journey Pattern](https://dzone.com/articles/page-objects-refactored-solid-steps-to-the-screenp) - by Antony Marcano, Andy Palmer, John Ferguson Smart, and Jan Molak
* 2016: [Screenplays and Journeys, Not Page Objects](https://testerstories.com/2016/06/screenplays-and-journeys-not-page-objects/) - blog post by Jeff Nyman
* 2016: [Screenplay Pattern](https://serenity-js.org/handbook/design/screenplay-pattern/) as described by Jan Molak
* 2016: [Serenity/JS](https://github.com/serenity-js/serenity-js) - Jan Molak starts the Serenity/JS project - the original JavaScript/TypeScript implementation of the Screenplay Pattern
* 2017: [Testing modern webapps. At scale.](https://www.slideshare.net/janmolak/testing-modern-webapps-at-scale) - Jan Molak introduces the idea of "Blended Testing"
* 2019: [ScreenPy](https://pypi.org/project/screenpy/) - Python implementation of the Screenplay Pattern by [Perry Goy](https://www.linkedin.com/in/perry-goy/)
* 2020: [Boa Constrictor](https://automationpanda.com/2020/10/16/introducing-boa-constrictor-the-net-screenplay-pattern/) - a .NET implementation of Screenplay by [Andrew Knight](https://automationpanda.com/)
* 2020: [Understanding Screenplay](https://cucumber.io/blog/bdd/understanding-screenplay-(part-1)/) - blog series by [Matt Wynne](https://blog.mattwynne.net/)
* 2021: [Cucumber Screenplay](https://github.com/cucumber/screenplay.js) and [Sub-second TDD](https://github.com/subsecondtdd/todo-subsecond) - implementation by [Aslak Hellesøy](https://www.aslakhellesoy.com/)
* 2021: [BDD in Action, 2nd Edition](https://www.manning.com/books/bdd-in-action-second-edition) by John Ferguson Smart and Jan Molak includes several chapters and many examples of using the Screenplay Pattern in Java and TypeScript
:::
