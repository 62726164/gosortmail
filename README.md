# gosortmail

gosortmail is a simple [MDA](https://en.wikipedia.org/wiki/Mail_delivery_agent) that sorts mail into [maildir](https://en.wikipedia.org/wiki/Maildir) folders based on user defined rules. It's written in Go and uses an XML configuration file.

Read more about why I wrote it [here]().

## gosortmailrc

gosortmailrc is the configuration and rules file. It must be located at /home/user/.gosortmailrc

To validate gosortmailrc, after editing, use xmllint:

```bash
xmllint --format gosortmailrc --output gosortmailrc
xmllint --schema gosortmailrc.xsd gosortmailrc --noout
```

## Example rule

```bash
<rule>
	<name>github</name>
	<section>subject</section>
	<contains>[github]</contains>
	<folder>/home/user/Mail/github</folder>
</rule>
```

## Notes

  * I wrote gosortmail to replace my simple [procmail](https://en.wikipedia.org/wiki/Procmail) email filtering rules.
