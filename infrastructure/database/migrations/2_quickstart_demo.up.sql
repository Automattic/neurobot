INSERT INTO "workflows" ("id","name","description","active") VALUES (1,'QuickStart Demo','This workflow is meant to show a quick demo',1);
INSERT INTO "triggers" ("id","name","description","variety","workflow_ids","active") VALUES (1,'CURL Request Catcher','This webhook trigger will receive your webhook request while showcasing the demo','webhook','1',1);
INSERT INTO "workflow_steps" ("id","name","description","variety","workflow_id","sort_order") VALUES (1,'Log to stdout','This workflow step will show the payload to stdout while showcasing the demo','stdout',1,0);
INSERT INTO "trigger_meta" ("id","trigger_id","key","value") VALUES (1,1,'urlSuffix','quickstart');
