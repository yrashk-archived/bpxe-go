# Code Reviews,
## And How to Get Paid for Doing Them

Producing high-quality software is an ongoing challenge. Proficiency,
good planning, tech stack and tooling all play a role in this process. However,
one of the unsung heroes is code reviews done by fellow software
engineers.

At this stage, BPXE is being developed by one person, will a possible
growth to a small team happening in the near future. However, our own biases,
myopia and allowances often stay in the way of ensuring the best possible
quality of created software. We don't know what we don't know, and we
often don't see things right in front of our eyes.

Thus, this is an *open invitation* to a wide audience of software engineers
to participate in this project by providing one-time, occasional or regular
**code reviews for a compensation**.

That's right — you can submit your code review and get paid for it.

## What kind of reviews?

There are two major categories of interest at the moment:

* Code & architecture quality
* Specification adherence

### Code & Architecture Quality

Doing this kind of reviews does not require you to know much or anything
about BPMN and adjacent standards & technologies, but it does require a certain
level of proficiency and experience with the tech stack used by the project
(mostly Go at this moment).

The type of concerns of interest, including but not limited to:

* Style/convention violations
* Areas to improve readability
* Logical errors
* Potential problems (safety, concurrency, deadlocks, panics, etc.)  
* Performance concerns
* Sporadically failing tests

### Specification Adherence

This one requires either prior knowledge of BPMN or willingness to learn
and comb through a [thick document](https://www.omg.org/spec/BPMN/2.0/PDF).

The issues of interest here lie within the realm of *divergence of behavior*
between the implementation and the specification (including when specification
is a bit too vague, but you reached a logical conclusion of what behavior
the software should exhibit), *missing details* (not whole missing features
as BPXE is still in its infancy) and *potential bugs* in implementing
intended behaviors.

This one likely requires deeper research than code & architecture quality
ones, but this also makes them so valuable.

## How much will I be paid?

That depends. At this moment, it is up to maintainer's determination. Minor/trivial
matters can bring you  a few dollars; something more concerning can be up to $100 or so;
and deep, important discoveries (especially ones with suggestions on how to improve or
fix the issue) can run upwards to $1,000 or so at this time. This is still a new process and
with more experience, the hope is to establish a clearer pay scale.

Payments will be made through PayPal, Wise, bank transfers and cryptocurrencies or
other mutually agreeable means, depending on the amount and borders involved. Payment
information will be exchanged privately by email or a private chat.

Maintainer keeps the right to deny compensation for reviews deemed nonsensical
or otherwise incomprehensible. If a previous review identified the same issue,
new reviews of the same will not be compensated.

The maintainer may offer additional perks, such
as recorded acknowledgements of contributions, leaderboards,
special chat roles or custom emojis, etc.

## How do I submit reviews?

Simply [open an issue](https://github.com/bpxe/bpxe/issues/new) and
assign the `code review` label to it. You'll hear back from the maintainer.

The reviews don't have to be done strictly on pull requests (in fact, these
tend to flow from sent to merged fairly quickly at the moment). Any piece
of code you find is up for grabs, provided it hasn't been replaced yet!

## Any guidelines?

* Take your time. Don't try to review everything at once or speed through
  the code. Commit messages in BPXE are typically written in a form of
  Problem/Solution statement and tend to carry some additional context.
* Provide constructive and clear explanation of identified issues. Clarity is
  of utmost value.
* Please dedicate one issue per identified concern, unless it is a group of
  minor or related concerns.
* Try to provide useful references if necessary (external documentation, links, other code, etc.)  
* If you want to provide regular reviews, consider subscribing to project's
  activity feed ("Watch All Activity").
* Feel free to ask questions -- either in [discussions](https://github.com/bpxe/bpxe/discussions/categories/code-reviews)
  or in a [chat group](https://discord.gg/yYWPnEb4FK).
  
## Is there a deadline?

No, it's an open process with no set deadline. That being said, the maintainer
reserves the right to terminate this process for any reason at any moment.

## Have more questions?

You can [start a discussion](https://github.com/bpxe/bpxe/discussions/categories/code-reviews)
right here on GitHub. You can also join our [Discord group](https://discord.gg/yYWPnEb4FK) and there's a
`#code-reviews` channel there. 